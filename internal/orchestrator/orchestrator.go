package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/HyppolitePrn/taskrunner/internal/report"
	"github.com/HyppolitePrn/taskrunner/internal/task"
)

const (
	defaultTimeout = 30 * time.Second
	defaultRetries = 0
)

type timeoutProvider interface {
	Timeout() time.Duration
}

type retryProvider interface {
	Retries() int
}

// Orchestrate runs tasks through a worker pool, applying WithVerbose (and any
// other Option) on top of the required workers count.
func Orchestrate(ctx context.Context, tasks []task.Task, workers int, opts ...Option) (report.Report, error) {
	cfg := &OrchestratorConfig{Workers: workers}
	for _, opt := range opts {
		opt(cfg)
	}

	validWorkers, _ := ValidateWorkers(cfg.Workers)
	cfg.Workers = validWorkers

	return run(ctx, tasks, cfg)
}

type job struct {
	index int
	task  task.Task
}

type outcome struct {
	index  int
	result report.TaskResult
}

func run(ctx context.Context, tasks []task.Task, cfg *OrchestratorConfig) (report.Report, error) {
	jobs := make(chan job)
	results := make(chan outcome, len(tasks))

	var wg sync.WaitGroup
	for range cfg.Workers {
		wg.Go(func() {
			for j := range jobs {
				results <- outcome{index: j.index, result: runTask(ctx, j.task, cfg.Verbose)}
			}
		})
	}

	go func() {
		defer close(jobs)
		for i, t := range tasks {
			select {
			case jobs <- job{index: i, task: t}:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	dispatched := make([]bool, len(tasks))
	taskResults := make([]report.TaskResult, len(tasks))
	for o := range results {
		taskResults[o.index] = o.result
		dispatched[o.index] = true
	}

	final := make([]report.TaskResult, 0, len(tasks))
	for i, ok := range dispatched {
		if ok {
			final = append(final, taskResults[i])
		}
	}

	return report.Report{Results: final}, nil
}

func runTask(ctx context.Context, t task.Task, verbose bool) report.TaskResult {
	timeout := taskTimeout(t)
	maxRetries := taskRetries(t)

	if verbose {
		_, _ = fmt.Fprintf(os.Stderr, "[%s] start\n", t.ID())
	}

	start := time.Now()
	var lastErr error
	attempts := 0

	for attempt := 0; attempt <= maxRetries; attempt++ {
		attempts++
		taskCtx, cancel := context.WithTimeout(ctx, timeout)
		lastErr = t.Execute(taskCtx)
		cancel()

		if lastErr == nil {
			break
		}
		if ctx.Err() != nil {
			break
		}
	}

	duration := time.Since(start)

	var status string
	switch {
	case lastErr == nil:
		status = report.StatusSuccess
	case errors.Is(lastErr, context.DeadlineExceeded):
		status = report.StatusTimeout
	default:
		status = report.StatusFailed
	}

	if verbose {
		_, _ = fmt.Fprintf(os.Stderr, "[%s] %s (%s)\n", t.ID(), status, duration)
	}

	return report.TaskResult{ID: t.ID(), Status: status, Duration: duration.String(), Attempts: attempts}
}

func taskTimeout(t task.Task) time.Duration {
	if tp, ok := t.(timeoutProvider); ok {
		return tp.Timeout()
	}
	return defaultTimeout
}

func taskRetries(t task.Task) int {
	if rp, ok := t.(retryProvider); ok {
		return rp.Retries()
	}
	return defaultRetries
}
