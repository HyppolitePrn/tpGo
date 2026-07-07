package task

import (
	"errors"
	"testing"
)

func TestTaskError_Error(t *testing.T) {
	err := &TaskError{Code: CodeExecution, TaskID: "t1", Err: errors.New("boom")}

	got := err.Error()
	want := "task t1: code 1: boom"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestTaskError_Unwrap(t *testing.T) {
	inner := errors.New("boom")
	err := &TaskError{Code: CodeExecution, TaskID: "t1", Err: inner}

	if !errors.Is(err, inner) {
		t.Errorf("errors.Is(err, inner) = false, want true")
	}
	if got := errors.Unwrap(err); got != inner {
		t.Errorf("Unwrap() = %v, want %v", got, inner)
	}
}
