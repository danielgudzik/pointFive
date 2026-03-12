// Package workerpool provides a generic concurrent worker-pool with a job store.
//
// Flow:
//
//	Submit(job) → input channel → [worker 1..N] → results stored in job
package workerpool

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Job is a generic batch of inputs and their processed outputs.
type Job[In, Out any] struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // pending | done
	Items     []In      `json:"items"`
	Results   []Out     `json:"results,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	DoneAt    time.Time `json:"done_at,omitempty"`
}

// Pipeline manages a worker pool and an in-memory job store.
type Pipeline[In, Out any] struct {
	workerCount int
	log         *slog.Logger
	processFn   func(context.Context, In) Out
	mu          sync.RWMutex
	jobs        map[string]*Job[In, Out]
}

// New creates a Pipeline that processes each input item using processFn.
func New[In, Out any](workerCount int, log *slog.Logger, processFn func(context.Context, In) Out) *Pipeline[In, Out] {
	if workerCount <= 0 {
		workerCount = 1
	}
	return &Pipeline[In, Out]{
		workerCount: workerCount,
		log:         log,
		processFn:   processFn,
		jobs:        make(map[string]*Job[In, Out]),
	}
}

// Submit enqueues a job and processes it asynchronously.
func (p *Pipeline[In, Out]) Submit(ctx context.Context, job *Job[In, Out]) {
	job.Status = "pending"
	job.CreatedAt = time.Now()

	p.mu.Lock()
	p.jobs[job.ID] = job
	p.mu.Unlock()

	go p.process(ctx, job)
}

// GetAll returns all jobs.
func (p *Pipeline[In, Out]) GetAll() []*Job[In, Out] {
	p.mu.RLock()
	defer p.mu.RUnlock()
	jobs := make([]*Job[In, Out], 0, len(p.jobs))
	for _, job := range p.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// Get retrieves a job by ID.
func (p *Pipeline[In, Out]) Get(id string) (*Job[In, Out], bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	j, ok := p.jobs[id]
	return j, ok
}

// process fans out items to workers, collects results, then marks the job done.
func (p *Pipeline[In, Out]) process(ctx context.Context, job *Job[In, Out]) {
	itemCh := make(chan In, len(job.Items))
	resultCh := make(chan Out, len(job.Items))

	// Fan out: start workers
	var wg sync.WaitGroup
	for range p.workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range itemCh {
				resultCh <- p.processFn(ctx, item)
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
	results := make([]Out, 0, len(job.Items))
	for r := range resultCh {
		results = append(results, r)
	}

	// Mark job done
	p.mu.Lock()
	job.Results = results
	job.Status = "done"
	job.DoneAt = time.Now()
	p.mu.Unlock()

	p.log.Info("job done", "id", job.ID, "items", len(results))
}
