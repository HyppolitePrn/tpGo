package orchestrator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/HyppolitePrn/taskrunner/internal/report"
	"github.com/HyppolitePrn/taskrunner/internal/task"
)

func scheduled(t task.Task, timeout time.Duration, retries int) task.Task {
	return &task.ScheduledTask{Task: t, TimeoutDuration: timeout, RetryCount: retries}
}

func TestOrchestrate_Success(t *testing.T) {
	tasks := []task.Task{
		scheduled(task.NewFakeTask("t1", task.BehaviorSuccess, time.Millisecond), time.Second, 0),
	}

	rep, err := Orchestrate(context.Background(), tasks, 2)
	if err != nil {
		t.Fatalf("Orchestrate() error = %v", err)
	}
	if len(rep.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(rep.Results))
	}
	if rep.Results[0].Status != report.StatusSuccess {
		t.Errorf("Status = %q, want %q", rep.Results[0].Status, report.StatusSuccess)
	}
	if rep.Results[0].Attempts != 1 {
		t.Errorf("Attempts = %d, want 1", rep.Results[0].Attempts)
	}
}

func TestOrchestrate_ResultsPreserveTaskOrder(t *testing.T) {
	tasks := []task.Task{
		scheduled(task.NewFakeTask("slow", task.BehaviorSuccess, 30*time.Millisecond), time.Second, 0),
		scheduled(task.NewFakeTask("fast", task.BehaviorSuccess, time.Millisecond), time.Second, 0),
	}

	rep, err := Orchestrate(context.Background(), tasks, 2)
	if err != nil {
		t.Fatalf("Orchestrate() error = %v", err)
	}
	if len(rep.Results) != 2 {
		t.Fatalf("len(Results) = %d, want 2", len(rep.Results))
	}
	if rep.Results[0].ID != "slow" || rep.Results[1].ID != "fast" {
		t.Errorf("Results order = [%s, %s], want [slow, fast]", rep.Results[0].ID, rep.Results[1].ID)
	}
}

func TestOrchestrate_RetryExhaustedFails(t *testing.T) {
	tasks := []task.Task{
		scheduled(task.NewFakeTask("t1", task.BehaviorFail, time.Millisecond), time.Second, 2),
	}

	rep, err := Orchestrate(context.Background(), tasks, 1)
	if err != nil {
		t.Fatalf("Orchestrate() error = %v", err)
	}
	if rep.Results[0].Status != report.StatusFailed {
		t.Errorf("Status = %q, want %q", rep.Results[0].Status, report.StatusFailed)
	}
	if rep.Results[0].Attempts != 3 {
		t.Errorf("Attempts = %d, want 3 (1 + 2 retries)", rep.Results[0].Attempts)
	}
}

func TestOrchestrate_RetryThenSucceed(t *testing.T) {
	flaky := &flakyTask{id: "t1", failures: 2}
	tasks := []task.Task{scheduled(flaky, time.Second, 2)}

	rep, err := Orchestrate(context.Background(), tasks, 1)
	if err != nil {
		t.Fatalf("Orchestrate() error = %v", err)
	}
	if rep.Results[0].Status != report.StatusSuccess {
		t.Errorf("Status = %q, want %q", rep.Results[0].Status, report.StatusSuccess)
	}
	if rep.Results[0].Attempts != 3 {
		t.Errorf("Attempts = %d, want 3", rep.Results[0].Attempts)
	}
}

func TestOrchestrate_Timeout(t *testing.T) {
	tasks := []task.Task{
		scheduled(task.NewFakeTask("t1", task.BehaviorTimeout, 5*time.Millisecond), 20*time.Millisecond, 0),
	}

	rep, err := Orchestrate(context.Background(), tasks, 1)
	if err != nil {
		t.Fatalf("Orchestrate() error = %v", err)
	}
	if rep.Results[0].Status != report.StatusTimeout {
		t.Errorf("Status = %q, want %q", rep.Results[0].Status, report.StatusTimeout)
	}
}

func TestOrchestrate_WorkerLimitRespected(t *testing.T) {
	const workers = 2
	var current, maxConcurrent int32
	var mu sync.Mutex

	tasks := make([]task.Task, 0, 6)
	for i := range 6 {
		tasks = append(tasks, scheduled(&trackingTask{
			id:      fmt.Sprintf("t%d", i),
			current: &current,
			max:     &maxConcurrent,
			mu:      &mu,
			delay:   20 * time.Millisecond,
		}, time.Second, 0))
	}

	_, err := Orchestrate(context.Background(), tasks, workers)
	if err != nil {
		t.Fatalf("Orchestrate() error = %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if maxConcurrent > int32(workers) {
		t.Errorf("max concurrent tasks = %d, want <= %d", maxConcurrent, workers)
	}
}

func TestOrchestrate_ParentCancelledPartialReport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	tasks := make([]task.Task, 0, 5)
	for i := range 5 {
		tasks = append(tasks, scheduled(task.NewFakeTask(fmt.Sprintf("t%d", i), task.BehaviorSuccess, 100*time.Millisecond), time.Second, 0))
	}

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	rep, err := Orchestrate(ctx, tasks, 1)
	if err != nil {
		t.Fatalf("Orchestrate() error = %v", err)
	}
	if len(rep.Results) >= len(tasks) {
		t.Errorf("len(Results) = %d, want < %d (partial report expected)", len(rep.Results), len(tasks))
	}
}

func TestOrchestrate_InvalidWorkersFallsBackToDefault(t *testing.T) {
	tasks := []task.Task{scheduled(task.NewFakeTask("t1", task.BehaviorSuccess, time.Millisecond), time.Second, 0)}

	rep, err := Orchestrate(context.Background(), tasks, 0)
	if err != nil {
		t.Fatalf("Orchestrate() error = %v", err)
	}
	if len(rep.Results) != 1 {
		t.Fatalf("len(Results) = %d, want 1", len(rep.Results))
	}
}

func TestOrchestrate_Verbose(t *testing.T) {
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = w

	tasks := []task.Task{scheduled(task.NewFakeTask("t1", task.BehaviorSuccess, time.Millisecond), time.Second, 0)}
	_, orchErr := Orchestrate(context.Background(), tasks, 1, WithVerbose(true))

	closeErr := w.Close()
	os.Stderr = oldStderr

	if orchErr != nil {
		t.Fatalf("Orchestrate() error = %v", orchErr)
	}
	if closeErr != nil {
		t.Fatalf("w.Close() error = %v", closeErr)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("io.Copy() error = %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "t1") {
		t.Errorf("verbose output missing task id, got: %q", out)
	}
	if !strings.Contains(out, "start") {
		t.Errorf("verbose output missing start marker, got: %q", out)
	}
	if !strings.Contains(out, report.StatusSuccess) {
		t.Errorf("verbose output missing success marker, got: %q", out)
	}
}

type flakyTask struct {
	id       string
	mu       sync.Mutex
	failures int
	calls    int
}

func (f *flakyTask) ID() string { return f.id }

func (f *flakyTask) Execute(_ context.Context) error {
	f.mu.Lock()
	f.calls++
	call := f.calls
	f.mu.Unlock()

	if call <= f.failures {
		return errors.New("flaky failure")
	}
	return nil
}

type trackingTask struct {
	id      string
	current *int32
	max     *int32
	mu      *sync.Mutex
	delay   time.Duration
}

func (t *trackingTask) ID() string { return t.id }

func (t *trackingTask) Execute(_ context.Context) error {
	n := atomic.AddInt32(t.current, 1)
	t.mu.Lock()
	if n > *t.max {
		*t.max = n
	}
	t.mu.Unlock()

	time.Sleep(t.delay)
	atomic.AddInt32(t.current, -1)
	return nil
}
