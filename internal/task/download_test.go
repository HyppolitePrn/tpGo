package task

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadTask_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "data.json")
	dt := NewDownloadTask("d1", srv.URL, dest)

	err := dt.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	content, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(content) != `{"ok":true}` {
		t.Errorf("content = %q, want %q", content, `{"ok":true}`)
	}
}

func TestDownloadTask_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "data.json")
	dt := NewDownloadTask("d1", srv.URL, dest)

	err := dt.Execute(context.Background())
	if err == nil {
		t.Fatal("Execute() error = nil, want error for 404 response")
	}

	var taskErr *TaskError
	if !errors.As(err, &taskErr) {
		t.Fatalf("Execute() error is not a *TaskError: %v", err)
	}
	if taskErr.Code != CodeExecution {
		t.Errorf("Code = %d, want %d", taskErr.Code, CodeExecution)
	}
}

func TestDownloadTask_ContextTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.Write([]byte("too late"))
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "data.json")
	dt := NewDownloadTask("d1", srv.URL, dest)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	err := dt.Execute(ctx)
	if err == nil {
		t.Fatal("Execute() error = nil, want timeout error")
	}

	var taskErr *TaskError
	if !errors.As(err, &taskErr) {
		t.Fatalf("Execute() error is not a *TaskError: %v", err)
	}
	if taskErr.Code != CodeTimeout {
		t.Errorf("Code = %d, want %d", taskErr.Code, CodeTimeout)
	}
}

func TestDownloadTask_InvalidURL(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "data.json")
	dt := NewDownloadTask("d1", "://not-a-url", dest)

	err := dt.Execute(context.Background())
	if err == nil {
		t.Fatal("Execute() error = nil, want error for invalid URL")
	}

	var taskErr *TaskError
	if !errors.As(err, &taskErr) {
		t.Fatalf("Execute() error is not a *TaskError: %v", err)
	}
	if taskErr.Code != CodeInvalidParams {
		t.Errorf("Code = %d, want %d", taskErr.Code, CodeInvalidParams)
	}
}
