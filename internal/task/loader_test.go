package task

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTasksFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tasks.json")
	err := os.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func TestLoadTasks_AllTypes(t *testing.T) {
	path := writeTasksFile(t, `{
		"tasks": [
			{"id": "t1", "type": "print", "params": {"message": "hi"}, "timeout": "2s", "retries": 0},
			{"id": "t2", "type": "download", "params": {"url": "https://example.com", "dest": "/tmp/x"}, "timeout": "5s", "retries": 2},
			{"id": "t3", "type": "calc", "params": {"value": 42}, "timeout": "1s", "retries": 1},
			{"id": "t4", "type": "fake", "params": {"behavior": "timeout", "delay": "3s"}, "timeout": "500ms", "retries": 1}
		]
	}`)

	tasks, err := LoadTasks(path)
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}
	if len(tasks) != 4 {
		t.Fatalf("len(tasks) = %d, want 4", len(tasks))
	}

	wantIDs := []string{"t1", "t2", "t3", "t4"}
	for i, want := range wantIDs {
		if tasks[i].ID() != want {
			t.Errorf("tasks[%d].ID() = %q, want %q", i, tasks[i].ID(), want)
		}
	}

	st, ok := tasks[3].(*ScheduledTask)
	if !ok {
		t.Fatalf("tasks[3] is not a *ScheduledTask")
	}
	if st.Timeout() != 500*time.Millisecond {
		t.Errorf("Timeout() = %v, want %v", st.Timeout(), 500*time.Millisecond)
	}
	if st.Retries() != 1 {
		t.Errorf("Retries() = %d, want 1", st.Retries())
	}
	if _, ok := st.Task.(*FakeTask); !ok {
		t.Errorf("underlying task is not a *FakeTask")
	}
}

func TestLoadTasks_FileNotFound(t *testing.T) {
	_, err := LoadTasks(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("LoadTasks() error = nil, want error for missing file")
	}
}

func TestLoadTasks_InvalidJSON(t *testing.T) {
	path := writeTasksFile(t, `not json`)

	_, err := LoadTasks(path)
	if err == nil {
		t.Fatal("LoadTasks() error = nil, want error for malformed JSON")
	}
}

func TestLoadTasks_UnknownType(t *testing.T) {
	path := writeTasksFile(t, `{"tasks": [{"id": "t1", "type": "unknown", "params": {}, "timeout": "1s", "retries": 0}]}`)

	_, err := LoadTasks(path)
	if err == nil {
		t.Fatal("LoadTasks() error = nil, want error for unknown type")
	}

	var taskErr *TaskError
	if !errors.As(err, &taskErr) {
		t.Fatalf("error is not a *TaskError: %v", err)
	}
	if taskErr.Code != CodeInvalidParams {
		t.Errorf("Code = %d, want %d", taskErr.Code, CodeInvalidParams)
	}
}

func TestLoadTasks_InvalidTimeout(t *testing.T) {
	path := writeTasksFile(t, `{"tasks": [{"id": "t1", "type": "print", "params": {"message": "hi"}, "timeout": "not-a-duration", "retries": 0}]}`)

	_, err := LoadTasks(path)
	if err == nil {
		t.Fatal("LoadTasks() error = nil, want error for invalid timeout")
	}
}

func TestLoadTasks_InvalidParams(t *testing.T) {
	path := writeTasksFile(t, `{"tasks": [{"id": "t1", "type": "calc", "params": "not-an-object", "timeout": "1s", "retries": 0}]}`)

	_, err := LoadTasks(path)
	if err == nil {
		t.Fatal("LoadTasks() error = nil, want error for invalid params")
	}
}

func TestLoadTasks_InvalidFakeBehavior(t *testing.T) {
	path := writeTasksFile(t, `{"tasks": [{"id": "t1", "type": "fake", "params": {"behavior": "bogus", "delay": "1ms"}, "timeout": "1s", "retries": 0}]}`)

	_, err := LoadTasks(path)
	if err == nil {
		t.Fatal("LoadTasks() error = nil, want error for invalid fake behavior")
	}
}
