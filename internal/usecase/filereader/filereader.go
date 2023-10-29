package filereader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/antonmisa/cliurlfetcher/internal/entity"
	"github.com/antonmisa/cliurlfetcher/internal/usecase"
	"github.com/antonmisa/cliurlfetcher/pkg/logger"
)

const (
	defaultShutdownTimeout = 100
	defaultMaxRetries      = 3
)

type FileReader struct {
	r               io.Reader
	logger          logger.Interface
	queue           usecase.QueueWriter
	ctx             context.Context
	wg              sync.WaitGroup
	shutdown        atomic.Bool
	shutdownTimeout time.Duration
}

var _ usecase.StartStoper = (*FileReader)(nil)

func New(ctx context.Context, r io.Reader, q usecase.QueueWriter, l logger.Interface) *FileReader {
	fr := &FileReader{
		ctx:             ctx,
		r:               r,
		logger:          l,
		queue:           q,
		shutdown:        atomic.Bool{},
		shutdownTimeout: defaultShutdownTimeout,
	}
	fr.shutdown.Store(false)

	return fr
}

func (fr *FileReader) Start() error {
	op := "FileReader - Start"

	fr.wg.Add(1)

	go func() {
		defer fr.wg.Done()

		fileScanner := bufio.NewScanner(fr.r)

		lineNumber := 1

		for fileScanner.Scan() {
			if fr.shutdown.Load() {
				return
			}

			select {
			case <-fr.ctx.Done():
				return
			default:
				url := fileScanner.Text()
				task := entity.Constructor(strconv.Itoa(lineNumber), url, defaultMaxRetries)

				err := fr.queue.Push(task)
				if err != nil {
					fr.logger.Error(" - fr.queue.Push: %w", op, err)
				}
			}

			lineNumber++
		}

		fr.logger.Info("%s input file read completed", op)
	}()

	return nil
}

// LazyShutdown -.
func (fr *FileReader) LazyShutdown() error {
	op := "FileReader - LazyShutdown"

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
func (fr *FileReader) Shutdown() error {
	op := "FileReader - Shutdown"

	fr.shutdown.Store(true)

	c := make(chan struct{})

	go func() {
		defer close(c)
		fr.wg.Wait()
	}()

	select {
	case <-c:
		return nil // completed normally
	case <-time.After(fr.shutdownTimeout):
		return fmt.Errorf("%s timed out", op) // timed out
	}
}
