package task

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type fileSpec struct {
	Tasks []taskSpec `json:"tasks"`
}

type taskSpec struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Params  json.RawMessage `json:"params"`
	Timeout string          `json:"timeout"`
	Retries int             `json:"retries"`
}

type printParams struct {
	Message string `json:"message"`
}

type downloadParams struct {
	URL  string `json:"url"`
	Dest string `json:"dest"`
}

type calcParams struct {
	Value float64 `json:"value"`
}

type fakeParams struct {
	Behavior string `json:"behavior"`
	Delay    string `json:"delay"`
}

func LoadTasks(path string) ([]Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load tasks: %w", err)
	}

	var file fileSpec
	err = json.Unmarshal(data, &file)
	if err != nil {
		return nil, fmt.Errorf("load tasks: parse %s: %w", path, err)
	}

	tasks := make([]Task, 0, len(file.Tasks))
	for _, spec := range file.Tasks {
		st, err := buildScheduledTask(spec)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, st)
	}
	return tasks, nil
}

func buildScheduledTask(spec taskSpec) (*ScheduledTask, error) {
	timeout, err := time.ParseDuration(spec.Timeout)
	if err != nil {
		return nil, &TaskError{Code: CodeInvalidParams, TaskID: spec.ID, Err: fmt.Errorf("invalid timeout %q: %w", spec.Timeout, err)}
	}

	base, err := newTaskByType(spec)
	if err != nil {
		return nil, err
	}

	return &ScheduledTask{Task: base, TimeoutDuration: timeout, RetryCount: spec.Retries}, nil
}

func newTaskByType(spec taskSpec) (Task, error) {
	switch spec.Type {
	case "print":
		var p printParams
		err := json.Unmarshal(spec.Params, &p)
		if err != nil {
			return nil, &TaskError{Code: CodeInvalidParams, TaskID: spec.ID, Err: err}
		}
		return NewPrintTask(spec.ID, p.Message), nil

	case "download":
		var p downloadParams
		err := json.Unmarshal(spec.Params, &p)
		if err != nil {
			return nil, &TaskError{Code: CodeInvalidParams, TaskID: spec.ID, Err: err}
		}
		return NewDownloadTask(spec.ID, p.URL, p.Dest), nil

	case "calc":
		var p calcParams
		err := json.Unmarshal(spec.Params, &p)
		if err != nil {
			return nil, &TaskError{Code: CodeInvalidParams, TaskID: spec.ID, Err: err}
		}
		return NewCalcTask(spec.ID, p.Value), nil

	case "fake":
		var p fakeParams
		err := json.Unmarshal(spec.Params, &p)
		if err != nil {
			return nil, &TaskError{Code: CodeInvalidParams, TaskID: spec.ID, Err: err}
		}

		behavior, err := parseFakeBehavior(p.Behavior)
		if err != nil {
			return nil, &TaskError{Code: CodeInvalidParams, TaskID: spec.ID, Err: err}
		}

		delay, err := time.ParseDuration(p.Delay)
		if err != nil {
			return nil, &TaskError{Code: CodeInvalidParams, TaskID: spec.ID, Err: fmt.Errorf("invalid delay %q: %w", p.Delay, err)}
		}
		return NewFakeTask(spec.ID, behavior, delay), nil

	default:
		return nil, &TaskError{Code: CodeInvalidParams, TaskID: spec.ID, Err: fmt.Errorf("unknown task type %q", spec.Type)}
	}
}

func parseFakeBehavior(s string) (FakeTaskBehavior, error) {
	switch s {
	case "success", "":
		return BehaviorSuccess, nil
	case "fail":
		return BehaviorFail, nil
	case "timeout":
		return BehaviorTimeout, nil
	default:
		return 0, fmt.Errorf("unknown fake behavior %q", s)
	}
}
