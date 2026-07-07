package task

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type DownloadTask struct {
	id     string
	url    string
	dest   string
	client *http.Client
}

func NewDownloadTask(id, url, dest string) *DownloadTask {
	return &DownloadTask{id: id, url: url, dest: dest, client: http.DefaultClient}
}

func (t *DownloadTask) ID() string { return t.id }

func (t *DownloadTask) Execute(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.url, nil)
	if err != nil {
		return &TaskError{Code: CodeInvalidParams, TaskID: t.id, Err: err}
	}

	resp, err := t.client.Do(req)
	if err != nil {
		code := CodeExecution
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			code = CodeTimeout
		}
		return &TaskError{Code: code, TaskID: t.id, Err: err}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return &TaskError{Code: CodeExecution, TaskID: t.id, Err: fmt.Errorf("unexpected status: %s", resp.Status)}
	}

	f, err := os.Create(t.dest)
	if err != nil {
		return &TaskError{Code: CodeExecution, TaskID: t.id, Err: err}
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return &TaskError{Code: CodeExecution, TaskID: t.id, Err: err}
	}
	return nil
}
