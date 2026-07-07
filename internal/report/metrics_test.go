package report

import (
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestWriteMetrics(t *testing.T) {
	results := []TaskResult{
		{ID: "t1", Status: StatusSuccess, Duration: "12ms", Attempts: 1},
		{ID: "t2", Status: StatusTimeout, Duration: "3.001s", Attempts: 3},
		{ID: "t3", Status: StatusFailed, Duration: "45ms", Attempts: 2},
		{ID: "t4", Status: StatusSuccess, Duration: "5ms", Attempts: 1},
	}

	md := WriteMetrics(results)

	wantLines := map[string]string{
		"Tâches exécutées":  "4",
		"Tâches réussies":   "2",
		"Tâches en échec":   "1",
		"Tâches en timeout": "1",
	}
	for label, want := range wantLines {
		line := findLine(t, md, label)
		got := strings.TrimSpace(strings.TrimPrefix(line, "- "+label+" : "))
		if got != want {
			t.Errorf("line %q = %q, want %q", label, got, want)
		}
	}

	goroutineLine := findLine(t, md, "Goroutines actives à la fin")
	goroutineStr := strings.TrimSpace(strings.TrimPrefix(goroutineLine, "- Goroutines actives à la fin : "))
	n, err := strconv.Atoi(goroutineStr)
	if err != nil {
		t.Fatalf("goroutine count is not an int: %q", goroutineStr)
	}
	if n < 1 || n > runtime.NumGoroutine()+5 {
		t.Errorf("goroutine count = %d, looks out of a sane range", n)
	}
}

func TestWriteMetrics_Empty(t *testing.T) {
	md := WriteMetrics(nil)

	line := findLine(t, md, "Tâches exécutées")
	got := strings.TrimSpace(strings.TrimPrefix(line, "- Tâches exécutées : "))
	if got != "0" {
		t.Errorf("Tâches exécutées = %q, want %q", got, "0")
	}
}

func findLine(t *testing.T, text, label string) string {
	t.Helper()
	for line := range strings.SplitSeq(text, "\n") {
		if strings.Contains(line, label) {
			return line
		}
	}
	t.Fatalf("no line containing %q found in:\n%s", label, text)
	return ""
}
