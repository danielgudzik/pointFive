// Package entities holds the core domain types shared across packages.
package entities

import "time"

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