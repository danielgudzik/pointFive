package entities

import "log/slog"

// PipelineSettings controls pipeline worker-pool behaviour.
type PipelineSettings struct {
	WorkerCount int
	Log         *slog.Logger
}
