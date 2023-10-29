package fetchprocessor

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/antonmisa/cliurlfetcher/internal/usecase"
	"github.com/antonmisa/cliurlfetcher/internal/usecase/fetcher"
	"github.com/antonmisa/cliurlfetcher/pkg/logger"
)

const (
	defaultShutdownTimeout time.Duration = 5 * time.Second

	defaultRetryWaitMinTime time.Duration = 50 * time.Millisecond
	defaultRetryWaitMaxTime time.Duration = 5 * time.Second
	defaultMaxAttempts      int           = 3
)

type FetchProcessor struct {
	workers         int
	logger          logger.Interface
	in              usecase.QueueReader
	out             usecase.QueueWriter
	ctx             context.Context
	wg              sync.WaitGroup
	shutdown        atomic.Bool
	shutdownTimeout time.Duration
}

var _ usecase.StartStoper = (*FetchProcessor)(nil)

func New(ctx context.Context, workers int, in usecase.QueueReader, out usecase.QueueWriter, l logger.Interface) *FetchProcessor {
	fr := &FetchProcessor{
		ctx:             ctx,
		workers:         workers,
		logger:          l,
		in:              in,
		out:             out,
		shutdown:        atomic.Bool{},
		shutdownTimeout: defaultShutdownTimeout,
	}
	fr.shutdown.Store(false)

	return fr
}

func (fr *FetchProcessor) Start() error {
	op := "FetchProcessor - Start"

	for i := 1; i <= fr.workers; i++ {
		fr.wg.Add(1)

		go func(id int) {
			defer fr.wg.Done()

			fr.logger.Info("%s number %d of %d", op, id, fr.workers)

			ftchr := fetcher.Constructor(fr.logger)

			for {
				if fr.shutdown.Load() {
					return
				}

				select {
				case <-fr.ctx.Done():
					return
				default:
					if task, ok := fr.in.Pop(); ok {
						if !task.IsReady() {
							fr.logger.Info("%s number %d received ready task with id %s", op, id, task.ID)

							err := fr.out.Push(task)
							if err != nil {
								fr.logger.Error(" - fr.out.Push: %w", op, err)
							}

							continue
						}

						req := fetcher.FetcherRequest{
							ID:     task.ID,
							Method: http.MethodGet,
							URL:    task.InputParams.URL,

							RetryWaitMin: defaultRetryWaitMinTime,
							RetryWaitMax: defaultRetryWaitMaxTime,
							MaxRetries:   task.CurrentState.MaxRetries,
						}

						task.OutputParams.TimeStarted = time.Now()

						resp, _ := ftchr.Get(fr.ctx, req)

						task.OutputParams.TimeCompleted = time.Now()

						task.CurrentState.Retries = resp.Retries
						task.OutputParams.StatusCode = resp.StatusCode
						task.OutputParams.Content = resp.Content
						task.OutputParams.ContentLength = resp.ContentLength

						err := fr.out.Push(task)
						if err != nil {
							fr.logger.Error(" - fr.out.Push: %w", op, err)
						}
					} else {
						fr.logger.Info("%s number %d no more data income", op, id)
						return
					}
				}
			}
		}(i)
	}

	return nil
}

// LazyShutdown -.
func (fr *FetchProcessor) LazyShutdown() error {
	op := "FetchProcessor - LazyShutdown"

	c := make(chan struct{})

	go func() {
		defer close(c)
		fr.wg.Wait()
	}()

	select {
	case <-c:
		return nil // completed normally
	case <-fr.ctx.Done():
		return fmt.Errorf("%s shutdown by context", op)
	}
}

// Shutdown -.
func (fr *FetchProcessor) Shutdown() error {
	op := "FetchProcessor - Shutdown"

	fr.shutdown.Store(true)

	c := make(chan struct{})

	go func() {
		defer close(c)
		fr.wg.Wait()
	}()

	select {
	case <-c:
		return nil // completed normally
	case <-fr.ctx.Done():
		return fmt.Errorf("%s shutdown by context", op)
	case <-time.After(fr.shutdownTimeout):
		return fmt.Errorf("%s timed out", op) // timed out
	}
}
