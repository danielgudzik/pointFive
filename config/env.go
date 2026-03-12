package config

// Env var key constants. Each name matches its string value exactly.
// Duration constants include _SECONDS to make the unit explicit.
const (
	SERVER_ADDR              = "SERVER_ADDR"
	WORKER_COUNT             = "WORKER_COUNT"
	READ_TIMEOUT_SECONDS     = "READ_TIMEOUT_SECONDS"
	WRITE_TIMEOUT_SECONDS    = "WRITE_TIMEOUT_SECONDS"
	SHUTDOWN_TIMEOUT_SECONDS = "SHUTDOWN_TIMEOUT_SECONDS"
	LOG_LEVEL                = "LOG_LEVEL"
)
