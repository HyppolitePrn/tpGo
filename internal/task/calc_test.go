package task

import (
	"context"
	"testing"
)

func TestCalcTask_Execute(t *testing.T) {
	ct := NewCalcTask("c1", 6)

	err := ct.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	if ct.Result != 36 {
		t.Errorf("Result = %v, want %v", ct.Result, 36)
	}
}

func TestCalcTask_Execute_CancelledContext(t *testing.T) {
	ct := NewCalcTask("c1", 6)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ct.Execute(ctx)
	if err == nil {
		t.Fatal("Execute() error = nil, want error for cancelled context")
	}
}
