package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/example/pointfive/entities"
	"github.com/example/pointfive/pipeline"
	"github.com/example/pointfive/utils/httputil"
)

type handlers struct {
	pipe *pipeline.ItemPipeline
	log  *slog.Logger
}

// GET /health
func (h *handlers) health(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// POST /jobs
// Body: { "items": [{ "id": "1", "payload": { "name": "alice" } }] }
func (h *handlers) submitJob(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Items []entities.Item `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if len(body.Items) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "items cannot be empty")
		return
	}

	job := &entities.ItemJob{
		ID:    httputil.NewID(),
		Items: body.Items,
	}

	h.pipe.Submit(job)

	httputil.WriteJSON(w, http.StatusAccepted, job)
}

// GET /jobs
func (h *handlers) listJobs(w http.ResponseWriter, r *http.Request) {
	status := r.PathValue("status")
	allJobs := h.pipe.GetAll()
	filteredJobs := make([]*entities.ItemJob, 0, len(allJobs))
	for _, job := range allJobs {
		if job.Status == status {
			filteredJobs = append(filteredJobs, job)
		}
	}
	httputil.WriteJSON(w, http.StatusOK, filteredJobs)
}

// DELETE /jobs/{id}
func (h *handlers) cancelJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	job, ok := h.pipe.Get(id)
	if !ok {
		httputil.WriteError(w, http.StatusNotFound, "job not found")
		return
	}
	if job.Status != "pending" {
		httputil.WriteError(w, http.StatusConflict, "job is not pending")
		return
	}

	h.pipe.Cancel(id)
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

// GET /jobs/{id}
func (h *handlers) getJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	job, ok := h.pipe.Get(id)
	if !ok {
		httputil.WriteError(w, http.StatusNotFound, "job not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, job)
}
