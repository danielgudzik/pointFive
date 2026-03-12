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
	Status    string    `json:"status"` // pending | done | cancelled
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
	cancels     map[string]context.CancelFunc
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
		cancels:     make(map[string]context.CancelFunc),
	}
}

// Submit enqueues a job and processes it asynchronously with its own independent context.
func (p *Pipeline[In, Out]) Submit(job *Job[In, Out]) {
	job.Status = "pending"
	job.CreatedAt = time.Now()

	jobCtx, cancel := context.WithCancel(context.Background())

	p.mu.Lock()
	p.jobs[job.ID] = job
	p.cancels[job.ID] = cancel
	p.mu.Unlock()

	go p.process(jobCtx, cancel, job)
}

// Cancel signals a job to stop processing. Returns false if the job is not found.
func (p *Pipeline[In, Out]) Cancel(id string) bool {
	p.mu.RLock()
	cancel, ok := p.cancels[id]
	p.mu.RUnlock()
	if !ok {
		return false
	}
	cancel()
	return true
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

// process fans out items to workers, collects results, then marks the job done or cancelled.
func (p *Pipeline[In, Out]) process(ctx context.Context, cancel context.CancelFunc, job *Job[In, Out]) {
	defer cancel()

	itemCh := make(chan In, len(job.Items))
	resultCh := make(chan Out, len(job.Items))

	// Fan out: start workers
	var wg sync.WaitGroup
	for range p.workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range itemCh {
				select {
				case <-ctx.Done():
					return
				default:
					resultCh <- p.processFn(ctx, item)
				}
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

	// Mark job done or cancelled
	p.mu.Lock()
	job.Results = results
	if ctx.Err() != nil {
		job.Status = "cancelled"
	} else {
		job.Status = "done"
	}
	job.DoneAt = time.Now()
	delete(p.cancels, job.ID)
	p.mu.Unlock()

	p.log.Info("job "+job.Status, "id", job.ID, "items", len(results))
}