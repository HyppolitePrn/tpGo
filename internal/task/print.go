package task

import (
	"context"
	"fmt"
)

type PrintTask struct {
	id      string
	message string
}

func NewPrintTask(id, message string) *PrintTask {
	return &PrintTask{id: id, message: message}
}

func (t *PrintTask) ID() string { return t.id }

func (t *PrintTask) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return &TaskError{Code: CodeTimeout, TaskID: t.id, Err: ctx.Err()}
	default:
	}
	fmt.Println(t.message)
	return nil
}
