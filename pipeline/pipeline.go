// Package pipeline processes data items concurrently using a worker pool.
//
// Flow:
//
//	Submit(job) → input channel → [worker 1..N] → results stored in job
package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/pointfive/entities"
)

// Pipeline manages the worker pool and job store.
type Pipeline struct {
	cfg  entities.PipelineSettings
	mu   sync.RWMutex
	jobs map[string]*entities.Job
}

func New(cfg entities.PipelineSettings) *Pipeline {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 1
	}
	return &Pipeline{
		cfg:  cfg,
		jobs: make(map[string]*entities.Job),
	}
}

// Submit enqueues a job and processes it asynchronously.
func (p *Pipeline) Submit(ctx context.Context, job *entities.Job) {
	job.Status = "pending"
	job.CreatedAt = time.Now()

	p.mu.Lock()
	p.jobs[job.ID] = job
	p.mu.Unlock()

	go p.process(ctx, job)
}

// GetAll returns all jobs.
func (p *Pipeline) GetAll() []*entities.Job {
	p.mu.RLock()
	defer p.mu.RUnlock()
	jobs := make([]*entities.Job, 0, len(p.jobs))
	for _, job := range p.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// Get retrieves a job by ID.
func (p *Pipeline) Get(id string) (*entities.Job, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	j, ok := p.jobs[id]
	return j, ok
}

// process fans out items to workers, collects results, then marks the job done.
func (p *Pipeline) process(ctx context.Context, job *entities.Job) {
	itemCh := make(chan entities.Item, len(job.Items))
	resultCh := make(chan entities.Result, len(job.Items))

	// Fan out: start workers
	var wg sync.WaitGroup
	for range p.cfg.WorkerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range itemCh {
				resultCh <- p.processItem(ctx, item)
			}
		}()
	}

	// Send items to workers
	for _, item := range job.Items {
		itemCh <- item
	}
	close(itemCh)

	// Wait for all workers, then close results
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	results := make([]entities.Result, 0, len(job.Items))
	for r := range resultCh {
		results = append(results, r)
	}

	// Mark job done
	p.mu.Lock()
	job.Results = results
	job.Status = "done"
	job.DoneAt = time.Now()
	p.mu.Unlock()

	p.cfg.Log.Info("job done", "id", job.ID, "items", len(results))
}

// processItem transforms a single item.
// ── ADD YOUR DATA PROCESSING LOGIC HERE ──
func (p *Pipeline) processItem(_ context.Context, item entities.Item) entities.Result {
	out := make(map[string]any, len(item.Payload))

	for k, v := range item.Payload {
		switch val := v.(type) {
		case string:
			out[k] = fmt.Sprintf("[processed] %s", val)
		case float64:
			out[k] = val * 2
		default:
			out[k] = v
		}
	}

	return entities.Result{ItemID: item.ID, Output: out}
}