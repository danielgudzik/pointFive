package pipeline

import (
	"context"
	"testing"

	"github.com/example/pointfive/entities"
)

func TestProcessTextUppercasesStrings(t *testing.T) {
	item := entities.Item{ID: "1", Type: "text", Payload: map[string]any{"message": "hello"}}

	result := processItem(context.Background(), item)

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	got, ok := result.Output["message"].(string)
	if !ok {
		t.Fatal("expected string output")
	}
	if got != "HELLO" {
		t.Errorf("got %q, want %q", got, "HELLO")
	}
}

func TestProcessMetricDoublesNumbers(t *testing.T) {
	item := entities.Item{ID: "2", Type: "metric", Payload: map[string]any{"cpu": float64(40)}}

	result := processItem(context.Background(), item)

	if result.Error != "" {
		t.Fatalf("unexpected error: %s", result.Error)
	}
	got, ok := result.Output["cpu"].(float64)
	if !ok {
		t.Fatal("expected float64 output")
	}
	if got != 80 {
		t.Errorf("got %v, want 80", got)
	}
}

func TestProcessItemUnknownTypeReturnsError(t *testing.T) {
	item := entities.Item{ID: "3", Type: "image", Payload: map[string]any{}}

	result := processItem(context.Background(), item)

	if result.Error == "" {
		t.Error("expected error for unknown type, got none")
	}
}