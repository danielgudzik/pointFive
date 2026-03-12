package pipeline

import (
	"context"
	"testing"

	"github.com/example/pointfive/entities"
)

func TestProcessItemTransformsStrings(t *testing.T) {
	item := entities.Item{ID: "x", Payload: map[string]any{"city": "NYC"}}

	result := processItem(context.Background(), item)

	got, ok := result.Output["city"].(string)
	if !ok {
		t.Fatal("expected string output for city")
	}
	if got != "[processed] NYC" {
		t.Errorf("got %q, want %q", got, "[processed] NYC")
	}
}

func TestProcessItemDoublesNumbers(t *testing.T) {
	item := entities.Item{ID: "x", Payload: map[string]any{"count": float64(5)}}

	result := processItem(context.Background(), item)

	got, ok := result.Output["count"].(float64)
	if !ok {
		t.Fatal("expected float64 output for count")
	}
	if got != 10 {
		t.Errorf("got %v, want 10", got)
	}
}
