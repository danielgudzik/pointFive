package pipeline

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/example/pointfive/entities"
)

func newTestPipeline() *Pipeline {
	return New(entities.PipelineSettings{
		WorkerCount: 2,
		Log:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
	})
}

func TestSubmitAndGet(t *testing.T) {
	p := newTestPipeline()

	job := &entities.Job{
		ID: "test-1",
		Items: []entities.Item{
			{ID: "a", Payload: map[string]any{"name": "alice", "score": float64(10)}},
			{ID: "b", Payload: map[string]any{"name": "bob"}},
		},
	}

	p.Submit(context.Background(), job)

	// Wait for job to complete (max 2s)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		j, ok := p.Get("test-1")
		if ok && j.Status == "done" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	got, ok := p.Get("test-1")
	if !ok {
		t.Fatal("job not found")
	}
	if got.Status != "done" {
		t.Fatalf("expected status done, got %s", got.Status)
	}
	if len(got.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got.Results))
	}
}

func TestProcessItemTransformsStrings(t *testing.T) {
	p := newTestPipeline()
	item := entities.Item{ID: "x", Payload: map[string]any{"city": "NYC"}}

	result := p.processItem(context.Background(), item)

	got, ok := result.Output["city"].(string)
	if !ok {
		t.Fatal("expected string output for city")
	}
	if got != "[processed] NYC" {
		t.Errorf("got %q, want %q", got, "[processed] NYC")
	}
}

func TestProcessItemDoublesNumbers(t *testing.T) {
	p := newTestPipeline()
	item := entities.Item{ID: "x", Payload: map[string]any{"count": float64(5)}}

	result := p.processItem(context.Background(), item)

	got, ok := result.Output["count"].(float64)
	if !ok {
		t.Fatal("expected float64 output for count")
	}
	if got != 10 {
		t.Errorf("got %v, want 10", got)
	}
}