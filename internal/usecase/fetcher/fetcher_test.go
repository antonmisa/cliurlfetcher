package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/antonmisa/cliurlfetcher/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestFetcher_Get(t *testing.T) {
	type args struct {
		ctx context.Context
		req FetcherRequest
	}
	tests := []struct {
		name    string
		f       func() Fetcher
		args    args
		want    FetcherResponse
		wantErr bool
	}{
		{
			name: "ok",
			f: func() Fetcher {
				l, _ := logger.NewFake()
				return Constructor(l)
			},
			args: args{
				ctx: context.Background(),
				req: FetcherRequest{
					ID:         "1",
					URL:        "http://www.blankwebsite.com/",
					Method:     http.MethodGet,
					MaxRetries: 3,
				},
			},
			want: FetcherResponse{
				ID:         "1",
				StatusCode: 200,
				Content:    "Blank website",
				ContentLength: 13,
				Retries:       1,
			},
		},
		{
			name: "404",
			f: func() Fetcher {
				l, _ := logger.NewFake()
				return Constructor(l)
			},
			args: args{
				ctx: context.Background(),
				req: FetcherRequest{
					ID:         "1",
					URL:        "http://www.blankwebsite.com/anton.svv",
					Method:     http.MethodGet,
					MaxRetries: 3,
				},
			},
			want: FetcherResponse{
				ID:            "1",
				StatusCode:    404,
				Content:       "404 - page not found",
				ContentLength: 20,
				Retries:       1,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(tc.want.StatusCode)
				res.Write([]byte(tc.want.Content))
			}))
			defer func() { testServer.Close() }()

			tc.args.req.URL = testServer.URL

			got, err := tc.f().Get(tc.args.ctx, tc.args.req)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, got, tc.want)
		})
	}
}
