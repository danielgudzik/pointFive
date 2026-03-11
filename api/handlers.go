package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/pointfive/pipeline"
)

type handlers struct {
	pipe *pipeline.Pipeline
	log  *slog.Logger
}

// GET /health
func (h *handlers) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /jobs
// Body: { "items": [{ "id": "1", "payload": { "name": "alice" } }] }
func (h *handlers) submitJob(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Items []pipeline.Item `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if len(body.Items) == 0 {
		writeError(w, http.StatusBadRequest, "items cannot be empty")
		return
	}

	job := &pipeline.Job{
		ID:    newID(),
		Items: body.Items,
	}

	h.pipe.Submit(r.Context(), job)

	writeJSON(w, http.StatusAccepted, job)
}

// GET /jobs/{id}
func (h *handlers) getJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	job, ok := h.pipe.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	writeJSON(w, http.StatusOK, job)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func newID() string {
	return time.Now().Format("20060102150405.999999999")
}
