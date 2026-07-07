package task

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFakeTask_Success(t *testing.T) {
	ft := NewFakeTask("f1", BehaviorSuccess, 10*time.Millisecond)

	err := ft.Execute(context.Background())
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

func TestFakeTask_Fail(t *testing.T) {
	ft := NewFakeTask("f1", BehaviorFail, 10*time.Millisecond)

	err := ft.Execute(context.Background())
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	var taskErr *TaskError
	if !errors.As(err, &taskErr) {
		t.Fatalf("Execute() error is not a *TaskError: %v", err)
	}
	if taskErr.Code != CodeExecution {
		t.Errorf("Code = %d, want %d", taskErr.Code, CodeExecution)
	}
}

func TestFakeTask_Timeout(t *testing.T) {
	ft := NewFakeTask("f1", BehaviorTimeout, 5*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := ft.Execute(ctx)
	if err == nil {
		t.Fatal("Execute() error = nil, want timeout error")
	}

	var taskErr *TaskError
	if !errors.As(err, &taskErr) {
		t.Fatalf("Execute() error is not a *TaskError: %v", err)
	}
	if taskErr.Code != CodeTimeout {
		t.Errorf("Code = %d, want %d", taskErr.Code, CodeTimeout)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("errors.Is(err, context.DeadlineExceeded) = false, want true")
	}
}

func TestFakeTask_ContextCancelledBeforeDelay(t *testing.T) {
	ft := NewFakeTask("f1", BehaviorSuccess, time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ft.Execute(ctx)
	if err == nil {
		t.Fatal("Execute() error = nil, want error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("errors.Is(err, context.Canceled) = false, want true")
	}
}
