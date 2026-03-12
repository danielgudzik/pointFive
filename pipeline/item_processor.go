package pipeline

import (
	"context"
	"fmt"

	"github.com/example/pointfive/entities"
)

// processItem transforms a single Item.
// ── ADD YOUR DATA PROCESSING LOGIC HERE ──
func processItem(_ context.Context, item entities.Item) entities.Result {
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
