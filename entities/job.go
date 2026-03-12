// Package entities holds the core domain types shared across packages.
package entities

import wp "github.com/example/pointfive/utils/workerpool"

// Item is a single unit of data to be processed.
// Type must be "text" or "metric"; unrecognised types are passed through unchanged.
type Item struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"` // "text" | "metric"
	Payload map[string]any `json:"payload"`
}

// Result holds the processed output of one Item.
type Result struct {
	ItemID string         `json:"item_id"`
	Output map[string]any `json:"output"`
	Error  string         `json:"error,omitempty"`
}

// ItemJob is a batch of Items submitted for processing.
type ItemJob = wp.Job[Item, Result]
