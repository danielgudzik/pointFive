package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/pointfive/api"
	"github.com/example/pointfive/config"
	"github.com/example/pointfive/entities"
	"github.com/example/pointfive/pipeline"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	pipe := pipeline.NewItemPipeline(entities.PipelineSettings{
		WorkerCount: cfg.WorkerCount,
		Log:         log,
	})

	srv := api.NewServer(api.Config{
		Addr:                   cfg.ServerAddr,
		Pipeline:               pipe,
		Log:                    log,
		ReadTimeoutSeconds:     cfg.ReadTimeoutSeconds,
		WriteTimeoutSeconds:    cfg.WriteTimeoutSeconds,
		ShutdownTimeoutSeconds: cfg.ShutdownTimeoutSeconds,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("server listening", "addr", cfg.ServerAddr)
	if err := srv.Run(ctx); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
