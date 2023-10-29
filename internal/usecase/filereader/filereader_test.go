package filereader

import (
	"context"
	"io"
	"testing"

	"github.com/antonmisa/cliurlfetcher/internal/entity"
	"github.com/antonmisa/cliurlfetcher/internal/usecase"
	"github.com/antonmisa/cliurlfetcher/internal/usecase/queue"
	"github.com/antonmisa/cliurlfetcher/pkg/logger"
	"github.com/stretchr/testify/require"
)

type HelperReader struct {
	pos int
	Buf []byte
}

func (hr *HelperReader) Read(p []byte) (int, error) {
	if hr.pos >= len(hr.Buf) {
		return 0, io.EOF
	}

	n := copy(p, hr.Buf)
	hr.pos += n
	return n, nil
}

func TestFileReader_Start(t *testing.T) {
	type ts struct {
		ok bool
		t  entity.Task
	}

	type rvs struct {
		err error
		ts  []ts
	}

	type args struct {
		r     HelperReader
		ctx   context.Context
		queue usecase.Queue
	}

	tests := []struct {
		name string
		args args
		fr   func(ctx context.Context, r io.Reader, qw usecase.QueueWriter) *FileReader
		rv   rvs
	}{
		{
			name: "ok",
			args: args{
				ctx: context.Background(),
				r: HelperReader{
					Buf: []byte("http://www.yandex.ru"),
				},
				queue: queue.New(),
			},
			fr: func(ctx context.Context, r io.Reader, qw usecase.QueueWriter) *FileReader {
				l, _ := logger.NewFake()
				return New(ctx, r, qw, l)
			},
			rv: rvs{
				err: nil,
				ts: []ts{
					{
						ok: true,
						t:  entity.Constructor("1", "http://www.yandex.ru", 3),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			fr := tc.fr(tc.args.ctx, &tc.args.r, tc.args.queue)
			err := fr.Start()
			if tc.rv.err != nil {
				require.Error(t, err, tc.rv.err)
			} else {
				require.NoError(t, err)
			}

			err = fr.LazyShutdown()
			require.NoError(t, err)

			tc.args.queue.Close()

			got, ok := tc.args.queue.Pop()
			require.Equal(t, ok, tc.rv.ts[0].ok)
			require.Equal(t, got, tc.rv.ts[0].t)
		})
	}
}
