package orchestrator

import "testing"

func TestWithWorkers(t *testing.T) {
	cfg := &OrchestratorConfig{}
	WithWorkers(7)(cfg)

	if cfg.Workers != 7 {
		t.Errorf("Workers = %d, want 7", cfg.Workers)
	}
}

func TestWithVerbose(t *testing.T) {
	cfg := &OrchestratorConfig{}
	WithVerbose(true)(cfg)

	if !cfg.Verbose {
		t.Errorf("Verbose = false, want true")
	}
}
