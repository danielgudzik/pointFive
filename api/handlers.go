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
	var body entities.SubmitJobRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if len(body.Items) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "items cannot be empty")
		return
	}

	items := make(map[string]entities.Item, len(body.Items))
	for _, item := range body.Items {
		items[item.ID] = item
	}
	job := &entities.ItemJob{
		ID:    httputil.NewID(),
		Items: items,
	}

	h.pipe.Submit(r.Context(), job)

	httputil.WriteJSON(w, http.StatusAccepted, job)
}

// GET /jobs
func (h *handlers) listJobs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	allJobs := h.pipe.GetAll()
	filteredJobs := make([]*entities.ItemJob, 0, len(allJobs))
	if status == "" {
		filteredJobs = append(filteredJobs, allJobs...)
	} else {
		for _, job := range allJobs {
			if job.Status == status {
				filteredJobs = append(filteredJobs, job)
			}
		}
	}

	httputil.WriteJSON(w, http.StatusOK, filteredJobs)
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

// POST /jobs/{id}/retry
func (h *handlers) retryJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	job, ok := h.pipe.Get(id)
	if !ok {
		httputil.WriteError(w, http.StatusNotFound, "job not found")
		return
	}
	if job.Status != "done" {
		httputil.WriteError(w, http.StatusConflict, "job is not done")
		return
	}

	failedItems := make(map[string]entities.Item, len(job.Items))
	// get all failed items
	for _, result := range job.Results {
		if result.Error != "" {
			failedItems[result.ItemID] = job.Items[result.ItemID]
		}
	}

	if len(failedItems) > 0 {
		newJob := &entities.ItemJob{
			ID:    httputil.NewID(),
			Items: failedItems,
		}

		h.pipe.Submit(r.Context(), newJob)
		httputil.WriteJSON(w, http.StatusAccepted, newJob)
		return
	}

	httputil.WriteError(w, http.StatusBadRequest, "no failed items to retry")
}
