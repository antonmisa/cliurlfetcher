package cli

import (
	"context"

	//	e "github.com/antonmisa/1cctl_cli/internal/common/clierror"
	"github.com/antonmisa/cliurlfetcher/internal/usecase"
	"github.com/antonmisa/cliurlfetcher/pkg/logger"
)

type CliCtrl struct {
	ctx    context.Context
	in     usecase.Queue
	out    usecase.Queue
	fr     usecase.StartStoper
	fw     usecase.StartStoper
	proc   usecase.StartStoper
	logger logger.Interface
}

func New(ctx context.Context, in usecase.Queue, out usecase.Queue, fr usecase.StartStoper, fw usecase.StartStoper, proc usecase.StartStoper, l logger.Interface) *CliCtrl {
	return &CliCtrl{
		ctx:    ctx,
		in:     in,
		out:    out,
		fr:     fr,
		fw:     fw,
		proc:   proc,
		logger: l,
	}
}

func (cc *CliCtrl) Start() {
	op := "CliCtrl - Start"

	// Start file Reader
	err := cc.fr.Start()
	if err != nil {
		cc.logger.Error("%s cc.fr.Start: %w", op, err)
		return
	}

	// Start output writer
	err = cc.fw.Start()
	if err != nil {
		cc.logger.Error("%s cc.fw.Start: %w", op, err)
		return
	}

	// Start workers
	err = cc.proc.Start()
	if err != nil {
		cc.logger.Error("%s cc.proc.Start: %w", op, err)
		return
	}

	// Wait for reader complete and close readed queue
	err = cc.fr.LazyShutdown()
	if err != nil {
		cc.logger.Error("%s cc.fr.LazyShutdown: %w", op, err)
	}

	cc.in.Close()

	// Wait for everyone completed
	err = cc.proc.LazyShutdown()
	if err != nil {
		cc.logger.Error("%s cc.proc.LazyShutdown: %w", op, err)
	}

	cc.out.Close()

	err = cc.fw.LazyShutdown()
	if err != nil {
		cc.logger.Error("%s cc.fw.LazyShutdown: %w", op, err)
	}
}
