package task

import (
	"context"
	"errors"
	"testing"
)

func TestPrintTask_Execute(t *testing.T) {
	pt := NewPrintTask("p1", "hello")

	if pt.ID() != "p1" {
		t.Errorf("ID() = %q, want %q", pt.ID(), "p1")
	}
	err := pt.Execute(context.Background())
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

func TestPrintTask_Execute_CancelledContext(t *testing.T) {
	pt := NewPrintTask("p1", "hello")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := pt.Execute(ctx)
	if err == nil {
		t.Fatal("Execute() error = nil, want error for cancelled context")
	}

	var taskErr *TaskError
	if !errors.As(err, &taskErr) {
		t.Fatalf("Execute() error is not a *TaskError: %v", err)
	}
	if taskErr.TaskID != "p1" {
		t.Errorf("TaskID = %q, want %q", taskErr.TaskID, "p1")
	}
}
