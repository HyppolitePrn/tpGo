package task

import "time"

// ScheduledTask wraps a Task with the per-task timeout and retry count
// read from the input file. The orchestrator type-asserts on Timeout()/Retries()
// to recover this scheduling metadata.
type ScheduledTask struct {
	Task
	TimeoutDuration time.Duration
	RetryCount      int
}

func (s *ScheduledTask) Timeout() time.Duration { return s.TimeoutDuration }

func (s *ScheduledTask) Retries() int { return s.RetryCount }
