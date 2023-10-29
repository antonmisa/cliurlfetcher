package fetcher

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/antonmisa/cliurlfetcher/pkg/logger"
	cleanhttp "github.com/hashicorp/go-cleanhttp"
)

const (
	defaultMaxIdleConns        = 100
	defaultMaxIdleConnsPerHost = 3
	defaultIdleConnTimeout     = 30

	defaultReadLimit = int64(128)
)

var (
	ErrNoMoreAttempts       = errors.New("no more attempts")
	ErrExternalRoutingError = errors.New("external or routing error")
)

type FetcherRequest struct {
	ID     string
	Method string
	URL    string

	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
	MaxRetries   int

	Request *http.Request
}

type FetcherResponse struct {
	ID            string
	StatusCode    int
	Content       string
	ContentLength int64
	Retries       int
}

// CheckRetry specifies a policy for handling retries. It is called
// following each request with the response and error values returned by
// the http.Client. If CheckRetry returns false, the Client stops retrying
// and returns the response to the caller. If CheckRetry returns an error,
// that error value is returned in lieu of the error from the request. The
// Client will close any response body when retrying, but if the retry is
// aborted it is up to the CheckRetry callback to properly close any
// response body before returning.
type CheckRetry func(ctx context.Context, resp *http.Response, err error) (bool, error)

// Backoff specifies a policy for how long to wait between retries.
// It is called after a failing request to determine the amount of time
// that should pass before trying again.
type Backoff func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration

type Fetcher struct {
	client *http.Client
	logger logger.Interface

	// CheckRetry specifies the policy for handling retries, and is called
	// after each request. The default policy is DefaultRetryPolicy.
	checkRetry CheckRetry

	// Backoff specifies the policy for how long to wait between retries
	backoff Backoff
}

// NewRequest creates a new wrapped request.
func NewRequest(ctx context.Context, method, url string) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, url, http.NoBody)
}

// DefaultRetryPolicy provides a default callback for Client.CheckRetry, which
// will retry on connection errors and server errors.
func DefaultRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	// don't propagate other errors
	shouldRetry := baseRetryPolicy(resp, err)
	return shouldRetry, err
}

func baseRetryPolicy(resp *http.Response, err error) bool {
	// If any error exist, than retryable too.
	if err != nil {
		return true
	}

	// 429 Too Many Requests is retryable. Sometimes the server puts
	// a Retry-After response header to indicate when the server is
	// available to start processing request from client.
	if resp.StatusCode == http.StatusTooManyRequests {
		return true
	}

	// 503 Service Unavailable is retryable. Sometimes the server puts
	// a Retry-After response header to indicate when the server is
	// available to start processing request from client.
	if resp.StatusCode == http.StatusServiceUnavailable {
		return true
	}

	return false
}

// DefaultBackoff provides a default callback for client.backoff which
// will perform exponential backoff based on the attempt number and limited
// by the provided minimum and maximum durations.
//
// It also tries to parse Retry-After response header when a http.StatusTooManyRequests
// (HTTP Code 429 or 503) is found in the resp parameter. Hence it will return the number of
// seconds the server states it may be ready to process more requests from this client.
func DefaultBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if resp != nil {
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			if s, ok := resp.Header["Retry-After"]; ok {
				if sleep, err := strconv.ParseInt(s[0], 10, 64); err == nil {
					return time.Second * time.Duration(sleep)
				}
			}
		}
	}

	mult := math.Pow(2, float64(attemptNum)) * float64(min)
	sleep := time.Duration(mult)
	if float64(sleep) != mult || sleep > max {
		sleep = max
	}
	return sleep
}

func Constructor(l logger.Interface) Fetcher {
	return Fetcher{
		client: cleanhttp.DefaultPooledClient(),
		logger: l,

		checkRetry: DefaultRetryPolicy,
		backoff:    DefaultBackoff,
	}
}

func (f Fetcher) Get(ctx context.Context, req FetcherRequest) (FetcherResponse, error) {
	op := "fetcher - Get"

	request, err := NewRequest(ctx, req.Method, req.URL)
	if err != nil {
		f.logger.Error("%s - NewRequest: %w", op, err)

		return FetcherResponse{
			ID: req.ID,
		}, err
	}

	req.Request = request

	return f.do(req)
}

func (f Fetcher) do(req FetcherRequest) (FetcherResponse, error) {
	op := "fetcher - DoWithContext"

	var lastStatusCode int

	var lastContentLength int64

	var attempt int

	for attempt = 1; attempt <= req.MaxRetries; attempt++ {
		f.logger.Info("%s - request %s starting attempt %d", op, req.ID, attempt)

		// Attempt the request
		resp, err := f.client.Do(req.Request)

		if resp != nil {
			lastStatusCode = resp.StatusCode
			lastContentLength = resp.ContentLength
		}

		// Check for retry if possible
		shouldRetry, err := f.checkRetry(req.Request.Context(), resp, err)
		if !shouldRetry || err != nil {
			f.logger.Info("%s - request %s stopped at attempt %d: %w", op, req.URL, attempt, err)

			var content string

			var errDrain error

			// stop and return request as-is
			if resp != nil {
				content, errDrain = f.drainBody(resp.Body)
			} else if err != nil {
				content = ErrExternalRoutingError.Error()
			}

			if errDrain != nil {
				f.logger.Error("%s - f.drainBody request %s: %w", op, req.ID, errDrain)

				return FetcherResponse{
					ID:         req.ID,
					StatusCode: lastStatusCode,
					Retries:    attempt,
				}, err
			}

			return FetcherResponse{
				ID:            req.ID,
				StatusCode:    lastStatusCode,
				Content:       content,
				ContentLength: lastContentLength,
				Retries:       attempt,
			}, err
		}

		f.drainBody(resp.Body)

		wait := f.backoff(req.RetryWaitMin, req.RetryWaitMax, attempt, resp)

		timer := time.NewTimer(wait)
		select {
		case <-req.Request.Context().Done():
			timer.Stop()

			f.client.CloseIdleConnections()

			return FetcherResponse{
				ID:            req.ID,
				StatusCode:    lastStatusCode,
				Retries:       attempt,
				ContentLength: lastContentLength,
			}, req.Request.Context().Err()
		case <-timer.C:
		}
	}

	// all attempts is gone, but nothing good happens
	return FetcherResponse{
		ID:            req.ID,
		StatusCode:    lastStatusCode,
		Retries:       attempt,
		ContentLength: lastContentLength,
	}, fmt.Errorf("%s - attempts is over for request %s: %w", op, req.ID, ErrNoMoreAttempts)
}

// Try to read the response body so we can reuse this connection.
func (f *Fetcher) drainBody(body io.ReadCloser) (string, error) {
	defer body.Close()

	op := "fetcher - drainBody"

	writer := bytes.NewBufferString("")

	_, err := io.Copy(writer, io.LimitReader(body, defaultReadLimit))
	if err != nil {
		f.logger.Error("%s - io.Copy: %w", op, err)
		return "", err
	}

	return writer.String(), err
}
