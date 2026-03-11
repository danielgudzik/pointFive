// Package pipeline processes data items concurrently using a worker pool.
//
// Flow:
//
//	Submit(job) → input channel → [worker 1..N] → results stored in job
package pipeline

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Item is a single unit of data to be processed.
type Item struct {
	ID      string         `json:"id"`
	Payload map[string]any `json:"payload"`
}

// Result holds the processed output of one Item.
type Result struct {
	ItemID string         `json:"item_id"`
	Output map[string]any `json:"output"`
	Error  string         `json:"error,omitempty"`
}

// Job is a batch of items submitted for processing.
type Job struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // pending | done
	Items     []Item    `json:"items"`
	Results   []Result  `json:"results,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	DoneAt    time.Time `json:"done_at,omitempty"`
}

// Config controls pipeline behaviour.
type Config struct {
	WorkerCount int
	Log         *slog.Logger
}

// Pipeline manages the worker pool and job store.
type Pipeline struct {
	cfg  Config
	mu   sync.RWMutex
	jobs map[string]*Job
}

func New(cfg Config) *Pipeline {
	if cfg.WorkerCount <= 0 {
		cfg.WorkerCount = 1
	}
	return &Pipeline{
		cfg:  cfg,
		jobs: make(map[string]*Job),
	}
}

// Submit enqueues a job and processes it asynchronously.
func (p *Pipeline) Submit(ctx context.Context, job *Job) {
	job.Status = "pending"
	job.CreatedAt = time.Now()

	p.mu.Lock()
	p.jobs[job.ID] = job
	p.mu.Unlock()

	go p.process(ctx, job)
}

// Get retrieves a job by ID.
func (p *Pipeline) Get(id string) (*Job, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	j, ok := p.jobs[id]
	return j, ok
}

// process fans out items to workers, collects results, then marks the job done.
func (p *Pipeline) process(ctx context.Context, job *Job) {
	itemCh := make(chan Item, len(job.Items))
	resultCh := make(chan Result, len(job.Items))

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
	results := make([]Result, 0, len(job.Items))
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
func (p *Pipeline) processItem(_ context.Context, item Item) Result {
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

	return Result{ItemID: item.ID, Output: out}
}
