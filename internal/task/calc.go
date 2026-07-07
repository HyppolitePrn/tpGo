package task

import "context"

type CalcTask struct {
	id     string
	value  float64
	Result float64
}

func NewCalcTask(id string, value float64) *CalcTask {
	return &CalcTask{id: id, value: value}
}

func (t *CalcTask) ID() string { return t.id }

func (t *CalcTask) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return &TaskError{Code: CodeTimeout, TaskID: t.id, Err: ctx.Err()}
	default:
	}
	t.Result = t.value * t.value
	return nil
}
