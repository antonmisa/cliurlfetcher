package entity

import "time"

type StateStatus int

const (
	StateStatusInitial StateStatus = iota
	StateStatusProcessing
	StateStatusCompleted
	StateStatusError
)

type State struct {
	Retries    int
	MaxRetries int
	Status     StateStatus
}

type InputParams struct {
	URL string
}

type OutputParams struct {
	StatusCode    int
	Content       string
	TimeStarted   time.Time
	TimeCompleted time.Time
	ContentLength int64
}

type Task struct {
	ID           string
	InputParams  InputParams
	OutputParams OutputParams
	CurrentState State
}

func Constructor(id, url string, maxRetries int) Task {
	return Task{
		ID: id,
		InputParams: InputParams{
			URL: url,
		},
		OutputParams: OutputParams{},
		CurrentState: State{
			Status:     StateStatusInitial,
			Retries:    0,
			MaxRetries: maxRetries,
		},
	}
}

func (t Task) IsReady() bool {
	if t.CurrentState.Status == StateStatusCompleted || t.CurrentState.Status == StateStatusError {
		return false
	}

	if t.CurrentState.Retries >= t.CurrentState.MaxRetries {
		return false
	}

	return true
}
