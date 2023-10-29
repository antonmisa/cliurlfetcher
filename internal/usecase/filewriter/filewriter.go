package filewriter

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/antonmisa/cliurlfetcher/internal/usecase"
	"github.com/antonmisa/cliurlfetcher/pkg/logger"
)

const (
	defaultShutdownTimeout = 100
)

type FileWriter struct {
	logger          logger.Interface
	sw              io.StringWriter
	queue           usecase.QueueReader
	ctx             context.Context
	wg              sync.WaitGroup
	shutdown        atomic.Bool
	shutdownTimeout time.Duration
}

var _ usecase.StartStoper = (*FileWriter)(nil)

func New(ctx context.Context, sw io.StringWriter, q usecase.QueueReader, l logger.Interface) *FileWriter {
	fr := &FileWriter{
		ctx:             ctx,
		sw:              sw,
		logger:          l,
		queue:           q,
		shutdown:        atomic.Bool{},
		shutdownTimeout: defaultShutdownTimeout,
	}
	fr.shutdown.Store(false)

	return fr
}

func (fw *FileWriter) Start() error {
	op := "FileWriter - Start"

	fw.wg.Add(1)

	go func() {
		defer fw.wg.Done()

		for {
			if fw.shutdown.Load() {
				return
			}

			select {
			case <-fw.ctx.Done():
				fw.logger.Info("%s output file write completed", op)

				return
			default:
				if task, ok := fw.queue.Pop(); ok {
					output := fmt.Sprintf("---------------\nCompleted url: %s, status: %d, contentlength: %d, content: %s\n", task.InputParams.URL, task.OutputParams.StatusCode, task.OutputParams.ContentLength, task.OutputParams.Content)

					_, err := fw.sw.WriteString(output)
					if err != nil {
						fw.logger.Error(" - fw.sw.WriteString: %w", op, err)
					}
				} else {
					fw.logger.Info("%s no more files income", op)

					return
				}
			}
		}
	}()

	return nil
}

// LazyShutdown -.
func (fw *FileWriter) LazyShutdown() error {
	op := "FileWriter - LazyShutdown"

	c := make(chan struct{})

	go func() {
		defer close(c)
		fw.wg.Wait()
	}()

	select {
	case <-c:
		fw.Done()
		return nil // completed normally
	case <-fw.ctx.Done():
		fw.Done()
		return fmt.Errorf("%s shutdown by context", op)
	}
}

// Shutdown -.
func (fw *FileWriter) Shutdown() error {
	op := "FileWriter - Shutdown"

	fw.shutdown.Store(true)

	c := make(chan struct{})

	go func() {
		defer close(c)
		fw.wg.Wait()
	}()

	select {
	case <-c:
		fw.Done()
		return nil // completed normally
	case <-time.After(fw.shutdownTimeout):
		fw.Done()
		return fmt.Errorf("%s timed out", op) // timed out
	}
}

// Done output -.
func (fw *FileWriter) Done() {
	op := "FileWriter - Done"

	output := "---------------\nDONE"

	_, err := fw.sw.WriteString(output)
	if err != nil {
		fw.logger.Error(" - fw.sw.WriteString: %w", op, err)
	}
}
