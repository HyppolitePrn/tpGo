package task

import "context"

type Task interface {
	ID() string
	Execute(ctx context.Context) error
}
