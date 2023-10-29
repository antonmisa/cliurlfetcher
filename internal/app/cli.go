package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antonmisa/cliurlfetcher/internal/config"
	cli "github.com/antonmisa/cliurlfetcher/internal/controller"
	"github.com/antonmisa/cliurlfetcher/internal/usecase/fetchprocessor"
	"github.com/antonmisa/cliurlfetcher/internal/usecase/filereader"
	"github.com/antonmisa/cliurlfetcher/internal/usecase/filewriter"
	"github.com/antonmisa/cliurlfetcher/internal/usecase/queue"
	"github.com/antonmisa/cliurlfetcher/pkg/logger"
)

func Run(cfg *config.Config, filePath string) {
	op := "app - Run"

	l, err := logger.New(cfg.Log.Path, cfg.Log.Level)
	if err != nil {
		l.Fatal(fmt.Errorf("%s - logger.New: %w", op, err))
	}

	fh, err := os.OpenFile(filePath, os.O_RDONLY, 0444)
	if err != nil {
		l.Fatal("%s could't read  file %s: %v", op, filePath, err)
	}
	defer fh.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		s := <-interrupt
		l.Info(fmt.Sprintf("%s - signal: %s", op, s.String()))
		cancel()
	}()

	in := queue.New()
	out := queue.New()

	fr := filereader.New(ctx, fh, in, l)
	fw := filewriter.New(ctx, os.Stdout, out, l)
	proc := fetchprocessor.New(ctx, cfg.NumberOfWorkers, in, out, l)

	ctrl := cli.New(ctx, in, out, fr, fw, proc, l)

	now := time.Now()

	l.Info(fmt.Sprintf("%s - started", op))

	ctrl.Start()

	l.Info("%s - succefully end, time taken: %s", op, time.Since(now).String())
}
