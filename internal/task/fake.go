package task

import (
	"context"
	"errors"
	"time"
)

type FakeTaskBehavior int

const (
	BehaviorSuccess FakeTaskBehavior = iota
	BehaviorFail
	BehaviorTimeout
)

type FakeTask struct {
	id       string
	behavior FakeTaskBehavior
	delay    time.Duration
}

func NewFakeTask(id string, behavior FakeTaskBehavior, delay time.Duration) *FakeTask {
	return &FakeTask{id: id, behavior: behavior, delay: delay}
}

func (t *FakeTask) ID() string { return t.id }

// BehaviorTimeout always waits for ctx cancellation after its delay, guaranteeing
// a timeout outcome regardless of the race between delay and the caller's deadline.
func (t *FakeTask) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return &TaskError{Code: CodeTimeout, TaskID: t.id, Err: ctx.Err()}
	case <-time.After(t.delay):
	}

	switch t.behavior {
	case BehaviorFail:
		return &TaskError{Code: CodeExecution, TaskID: t.id, Err: errors.New("fake task failed")}
	case BehaviorTimeout:
		<-ctx.Done()
		return &TaskError{Code: CodeTimeout, TaskID: t.id, Err: ctx.Err()}
	default:
		return nil
	}
}
