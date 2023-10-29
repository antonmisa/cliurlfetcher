package queue

import (
	"testing"

	"github.com/antonmisa/cliurlfetcher/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestTaskQueue_Push(t *testing.T) {
	type args struct {
		t entity.Task
	}

	type rvs struct {
		err error
	}

	tests := []struct {
		name string
		tq   *TaskQueue
		wrk  func(q *TaskQueue)
		args args
		rv   rvs
	}{
		{
			"ok push",
			New(),
			func(q *TaskQueue) {},
			args{
				entity.Task{
					ID: "1",
				},
			},
			rvs{
				nil,
			},
		},
		{
			"error push on closed",
			New(),
			func(q *TaskQueue) { q.Close() },
			args{
				entity.Task{
					ID: "1",
				},
			},
			rvs{
				ErrTaskQueueClosed,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.wrk(tc.tq)
			err := tc.tq.Push(tc.args.t)
			if tc.rv.err != nil {
				require.Error(t, err, tc.rv.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTaskQueue_Pop(t *testing.T) {
	type args struct {
		t entity.Task
	}

	type rvs struct {
		err error
		ok  bool
		t   entity.Task
	}

	tests := []struct {
		name string
		tq   *TaskQueue
		wrk  func(q *TaskQueue, t entity.Task)
		args args
		rv   rvs
	}{
		{
			"ok push-pop",
			New(),
			func(q *TaskQueue, t entity.Task) {
				q.Push(t)
			},
			args{
				entity.Task{
					ID: "1",
				},
			},
			rvs{
				nil,
				true,
				entity.Task{
					ID: "1",
				},
			},
		},
		{
			"pop on empty closed",
			New(),
			func(q *TaskQueue, t entity.Task) { q.Close() },
			args{
				entity.Task{
					ID: "1",
				},
			},
			rvs{
				nil,
				false,
				entity.Task{},
			},
		},
		{
			"pop on non empty closed",
			New(),
			func(q *TaskQueue, t entity.Task) {
				q.Push(t)
				q.Close()
			},
			args{
				entity.Task{
					ID: "1",
				},
			},
			rvs{
				nil,
				true,
				entity.Task{
					ID: "1",
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tc.wrk(tc.tq, tc.args.t)
			task, ok := tc.tq.Pop()
			require.Equal(t, ok, tc.rv.ok)
			require.Equal(t, task, tc.rv.t)
		})
	}
}
