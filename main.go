package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/pointfive/api"
	"github.com/example/pointfive/pipeline"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Pipeline: the data processing engine
	// Change workerCount to control parallelism
	pipe := pipeline.New(pipeline.Config{
		WorkerCount: 4,
		Log:         log,
	})

	// API server: exposes the pipeline over HTTP
	srv := api.NewServer(api.Config{
		Addr:     ":8080",
		Pipeline: pipe,
		Log:      log,
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("server listening", "addr", ":8080")
	if err := srv.Run(ctx); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
