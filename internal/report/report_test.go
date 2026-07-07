package report

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestReport_WriteTo(t *testing.T) {
	r := Report{Results: []TaskResult{
		{ID: "t1", Status: StatusSuccess, Duration: "12ms", Attempts: 1},
		{ID: "t2", Status: StatusTimeout, Duration: "3.001s", Attempts: 3},
	}}

	var buf bytes.Buffer
	n, err := r.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}
	if n != int64(buf.Len()) {
		t.Errorf("n = %d, want %d", n, buf.Len())
	}

	var got Report
	err = json.Unmarshal(buf.Bytes(), &got)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(got.Results) != 2 {
		t.Fatalf("len(got.Results) = %d, want 2", len(got.Results))
	}
	if got.Results[0] != r.Results[0] {
		t.Errorf("Results[0] = %+v, want %+v", got.Results[0], r.Results[0])
	}
	if got.Results[1] != r.Results[1] {
		t.Errorf("Results[1] = %+v, want %+v", got.Results[1], r.Results[1])
	}
}

func TestReport_WriteTo_Empty(t *testing.T) {
	r := Report{}

	var buf bytes.Buffer
	_, err := r.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	var got Report
	err = json.Unmarshal(buf.Bytes(), &got)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if len(got.Results) != 0 {
		t.Errorf("len(got.Results) = %d, want 0", len(got.Results))
	}
}
