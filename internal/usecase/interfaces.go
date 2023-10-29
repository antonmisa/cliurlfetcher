// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"github.com/antonmisa/cliurlfetcher/internal/entity"
)

//go:generate go run github.com/vektra/mockery/v2@v2.32.0 --all

type (
	// Queue -.
	Queue interface {
		Close()

		QueueReader
		QueueWriter
	}

	// QueueReader -.
	QueueReader interface {
		Pop() (entity.Task, bool)
	}

	// QueueWriter -.
	QueueWriter interface {
		Push(entity.Task) error
	}

	// StartShutdowner -.
	StartStoper interface {
		Start() error
		Shutdown() error
		LazyShutdown() error
	}
)
