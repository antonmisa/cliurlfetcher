package filewriter

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/antonmisa/cliurlfetcher/internal/entity"
	"github.com/antonmisa/cliurlfetcher/internal/usecase"
	"github.com/antonmisa/cliurlfetcher/internal/usecase/mocks"
	"github.com/antonmisa/cliurlfetcher/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestFileWriter_Start(t *testing.T) {
	type rvs struct {
		err    error
		output string
	}

	type args struct {
		f   strings.Builder
		ctx context.Context
	}

	tests := []struct {
		name     string
		args     args
		fr       func(ctx context.Context, f io.StringWriter, qr usecase.QueueReader) *FileWriter
		mockTask entity.Task
		mockOk   bool
		rv       rvs
	}{
		{
			name: "ok",
			args: args{
				ctx: context.Background(),
				f:   strings.Builder{},
			},
			fr: func(ctx context.Context, f io.StringWriter, qr usecase.QueueReader) *FileWriter {
				l, _ := logger.NewFake()
				return New(ctx, f, qr, l)
			},
			mockTask: entity.Constructor("1", "http://www.yandex.ru", 3),
			mockOk:   true,
			rv: rvs{
				err:    nil,
				output: "---------------\nCompleted url: http://www.yandex.ru, status: 0, contentlength: 0, content: \n---------------\nDONE",
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			queueMock := mocks.NewQueue(t)

			queueMock.On("Pop").
				Return(tc.mockTask, tc.mockOk).Once()

			queueMock.On("Pop").
				Return(entity.Task{}, false).Once()

			fr := tc.fr(tc.args.ctx, &tc.args.f, queueMock)
			err := fr.Start()
			if tc.rv.err != nil {
				require.Error(t, err, tc.rv.err)
			} else {
				require.NoError(t, err)
			}

			err = fr.LazyShutdown()
			require.NoError(t, err)

			require.Equal(t, tc.args.f.String(), tc.rv.output)
		})
	}
}
