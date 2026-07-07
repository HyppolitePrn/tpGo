package orchestrator

type OrchestratorConfig struct {
	Workers int
	Verbose bool
}

type Option func(*OrchestratorConfig)

func WithWorkers(n int) Option {
	return func(c *OrchestratorConfig) { c.Workers = n }
}

func WithVerbose(v bool) Option {
	return func(c *OrchestratorConfig) { c.Verbose = v }
}
