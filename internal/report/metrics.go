package report

import (
	"fmt"
	"runtime"
	"strings"
)

func WriteMetrics(results []TaskResult) string {
	var success, failed, timeout int
	for _, r := range results {
		switch r.Status {
		case StatusSuccess:
			success++
		case StatusFailed:
			failed++
		case StatusTimeout:
			timeout++
		}
	}

	var b strings.Builder
	b.WriteString("# Métriques d'exécution\n\n")
	_, _ = fmt.Fprintf(&b, "- Goroutines actives à la fin : %d\n", runtime.NumGoroutine())
	_, _ = fmt.Fprintf(&b, "- Tâches exécutées : %d\n", len(results))
	_, _ = fmt.Fprintf(&b, "- Tâches réussies : %d\n", success)
	_, _ = fmt.Fprintf(&b, "- Tâches en échec : %d\n", failed)
	_, _ = fmt.Fprintf(&b, "- Tâches en timeout : %d\n", timeout)
	return b.String()
}
