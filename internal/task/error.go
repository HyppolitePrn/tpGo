package task

import "fmt"

const (
	CodeExecution = iota + 1
	CodeTimeout
	CodeInvalidParams
)

type TaskError struct {
	Code   int
	TaskID string
	Err    error
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("task %s: code %d: %v", e.TaskID, e.Code, e.Err)
}

func (e *TaskError) Unwrap() error {
	return e.Err
}
