package logger

import (
	"os"
)

// FakeLogger -.
type FakeLogger struct {
}

var _ Interface = (*Logger)(nil)

// New -.
func NewFake() (*FakeLogger, error) {
	return &FakeLogger{}, nil
}

// Debug -.
func (l *FakeLogger) Debug(message interface{}, args ...interface{}) {
	return
}

// Info -.
func (l *FakeLogger) Info(message string, args ...interface{}) {
	return
}

// Warn -.
func (l *FakeLogger) Warn(message string, args ...interface{}) {
	return
}

// Error -.
func (l *FakeLogger) Error(message interface{}, args ...interface{}) {
	return
}

// Fatal -.
func (l *FakeLogger) Fatal(message interface{}, args ...interface{}) {
	os.Exit(1)
}

func (l *FakeLogger) log(message string, args ...interface{}) {
	return
}
