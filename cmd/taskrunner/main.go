package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/HyppolitePrn/taskrunner/internal/orchestrator"
	"github.com/HyppolitePrn/taskrunner/internal/report"
	"github.com/HyppolitePrn/taskrunner/internal/task"
)

func main() {
	filePath := flag.String("file", "", "path to tasks.json (required)")
	workers := flag.Int("workers", 3, "number of concurrent workers")
	verbose := flag.Bool("verbose", false, "log task status to stderr")
	flag.Parse()

	if *filePath == "" {
		_, _ = fmt.Fprintln(os.Stderr, "error: -file is required")
		os.Exit(1)
	}

	err := run(context.Background(), *filePath, *workers, *verbose)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, filePath string, workers int, verbose bool) error {
	validWorkers, err := orchestrator.ValidateWorkers(workers)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "warning: %v, falling back to %d workers\n", err, validWorkers)
	}

	tasks, err := task.LoadTasks(filePath)
	if err != nil {
		return fmt.Errorf("load tasks: %w", err)
	}

	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	rep, orchErr := orchestrator.Orchestrate(shutdownCtx, tasks, validWorkers, orchestrator.WithVerbose(verbose))

	reportErr := writeReport(rep)
	metricsErr := writeMetrics(rep)

	if orchErr != nil {
		return fmt.Errorf("orchestrate: %w", orchErr)
	}
	if reportErr != nil {
		return reportErr
	}
	return metricsErr
}

func writeReport(rep report.Report) error {
	_, err := rep.WriteTo(os.Stdout)
	if err != nil {
		return fmt.Errorf("write report: %w", err)
	}
	return nil
}

func writeMetrics(rep report.Report) error {
	content := report.WriteMetrics(rep.Results)

	err := os.WriteFile("METRICS.md", []byte(content), 0o644)
	if err != nil {
		return fmt.Errorf("write metrics: %w", err)
	}
	return nil
}
