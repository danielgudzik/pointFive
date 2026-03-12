package config

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

// AppConfig is the single source of truth for all runtime configuration.
type AppConfig struct {
	ServerAddr             string
	WorkerCount            int
	ReadTimeoutSeconds     int
	WriteTimeoutSeconds    int
	ShutdownTimeoutSeconds int
	LogLevel               slog.Level
}

// Load reads configuration in priority order:
//  1. Environment variables
//  2. .env file at the project root
//  3. Compiled-in defaults
func Load() (*AppConfig, error) {
	v := viper.New()

	v.SetDefault(SERVER_ADDR, ":8080")
	v.SetDefault(WORKER_COUNT, 4)
	v.SetDefault(READ_TIMEOUT_SECONDS, 10)
	v.SetDefault(WRITE_TIMEOUT_SECONDS, 30)
	v.SetDefault(SHUTDOWN_TIMEOUT_SECONDS, 10)
	v.SetDefault(LOG_LEVEL, "info")

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("config: read .env: %w", err)
		}
	}

	v.AutomaticEnv() // real env vars take precedence over .env

	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(v.GetString(LOG_LEVEL))); err != nil {
		return nil, fmt.Errorf("config: %s: %w", LOG_LEVEL, err)
	}

	return &AppConfig{
		ServerAddr:             v.GetString(SERVER_ADDR),
		WorkerCount:            v.GetInt(WORKER_COUNT),
		ReadTimeoutSeconds:     v.GetInt(READ_TIMEOUT_SECONDS),
		WriteTimeoutSeconds:    v.GetInt(WRITE_TIMEOUT_SECONDS),
		ShutdownTimeoutSeconds: v.GetInt(SHUTDOWN_TIMEOUT_SECONDS),
		LogLevel:               lvl,
	}, nil
}
