package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/pointfive/pipeline"
)

// Config holds server dependencies.
type Config struct {
	Addr     string
	Pipeline *pipeline.Pipeline
	Log      *slog.Logger
}

// Server wraps http.Server with a clean Run method.
type Server struct {
	http *http.Server
	log  *slog.Logger
}

// NewServer wires routes and returns a ready-to-run server.
func NewServer(cfg Config) *Server {
	mux := http.NewServeMux()
	h := &handlers{pipe: cfg.Pipeline, log: cfg.Log}

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /jobs", h.submitJob)
	mux.HandleFunc("GET /jobs/{id}", h.getJob)

	return &Server{
		http: &http.Server{
			Addr:         cfg.Addr,
			Handler:      withLogging(mux, cfg.Log),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		log: cfg.Log,
	}
}

// Run starts the server and shuts it down cleanly when ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	errc := make(chan error, 1)
	go func() { errc <- s.http.ListenAndServe() }()

	select {
	case <-ctx.Done():
		s.log.Info("shutting down")
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.http.Shutdown(shutCtx)
	case err := <-errc:
		return err
	}
}

func withLogging(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Info("←", "method", r.Method, "path", r.URL.Path, "took", time.Since(start))
	})
}
