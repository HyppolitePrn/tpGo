package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w

	fn()

	closeErr := w.Close()
	os.Stdout = oldStdout
	if closeErr != nil {
		t.Fatalf("w.Close() error = %v", closeErr)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("io.Copy() error = %v", err)
	}
	return buf.String()
}

func TestRun_Success(t *testing.T) {
	t.Chdir(t.TempDir())
	tasksPath := writeTasksFile(t, `{"tasks": [
		{"id": "t1", "type": "fake", "params": {"behavior": "success", "delay": "1ms"}, "timeout": "1s", "retries": 0},
		{"id": "t2", "type": "fake", "params": {"behavior": "fail", "delay": "1ms"}, "timeout": "1s", "retries": 1}
	]}`)

	var runErr error
	stdout := captureStdout(t, func() {
		runErr = run(context.Background(), tasksPath, 2, false)
	})
	if runErr != nil {
		t.Fatalf("run() error = %v", runErr)
	}

	if !strings.Contains(stdout, `"t1"`) || !strings.Contains(stdout, `"t2"`) {
		t.Errorf("stdout report missing task ids, got: %q", stdout)
	}

	metrics, err := os.ReadFile("METRICS.md")
	if err != nil {
		t.Fatalf("ReadFile(METRICS.md) error = %v", err)
	}
	if !strings.Contains(string(metrics), "Tâches exécutées : 2") {
		t.Errorf("METRICS.md missing expected content, got: %q", metrics)
	}
}

func TestRun_InvalidFile(t *testing.T) {
	t.Chdir(t.TempDir())

	err := run(context.Background(), "does-not-exist.json", 2, false)
	if err == nil {
		t.Fatal("run() error = nil, want error for missing file")
	}
	if _, statErr := os.Stat("METRICS.md"); statErr == nil {
		t.Error("METRICS.md was written despite load failure")
	}
}

func TestRun_InvalidWorkersFallsBack(t *testing.T) {
	t.Chdir(t.TempDir())
	tasksPath := writeTasksFile(t, `{"tasks": [
		{"id": "t1", "type": "fake", "params": {"behavior": "success", "delay": "1ms"}, "timeout": "1s", "retries": 0}
	]}`)

	var runErr error
	captureStdout(t, func() {
		runErr = run(context.Background(), tasksPath, 0, false)
	})
	if runErr != nil {
		t.Fatalf("run() error = %v", runErr)
	}
}

func TestRun_CancelledContextWritesPartialReport(t *testing.T) {
	t.Chdir(t.TempDir())
	tasksPath := writeTasksFile(t, `{"tasks": [
		{"id": "t1", "type": "fake", "params": {"behavior": "success", "delay": "1ms"}, "timeout": "1s", "retries": 0}
	]}`)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var runErr error
	captureStdout(t, func() {
		runErr = run(ctx, tasksPath, 1, false)
	})
	if runErr != nil {
		t.Fatalf("run() error = %v", runErr)
	}

	if _, statErr := os.Stat("METRICS.md"); statErr != nil {
		t.Errorf("METRICS.md should still be written on cancellation, stat error = %v", statErr)
	}
}
