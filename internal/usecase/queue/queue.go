package queue

import (
	"errors"

	"github.com/antonmisa/cliurlfetcher/internal/entity"
	"github.com/antonmisa/cliurlfetcher/internal/usecase"
)

// TaskQueue -.
type TaskQueue struct {
	ch     chan entity.Task
	closed bool
}

var ErrTaskQueueClosed = errors.New("task queue is closed")

var _ usecase.QueueWriter = (*TaskQueue)(nil)
var _ usecase.QueueReader = (*TaskQueue)(nil)
var _ usecase.Queue = (*TaskQueue)(nil)

func New() *TaskQueue {
	tq := &TaskQueue{
		ch: make(chan entity.Task, 1),
	}

	return tq
}

// Push -.
func (tq *TaskQueue) Push(t entity.Task) error {
	if tq.closed {
		return ErrTaskQueueClosed
	}

	tq.ch <- t

	return nil
}

// Pop -.
func (tq *TaskQueue) Pop() (entity.Task, bool) {
	v, ok := <-tq.ch
	return v, ok
}

// Close -.
func (tq *TaskQueue) Close() {
	close(tq.ch)
	tq.closed = true
}
