package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/example/pointfive/entities"
)

// processItem dispatches to the correct processor based on item.Type.
func processItem(ctx context.Context, item entities.Item) entities.Result {
	switch item.Type {
	case "text":
		return processText(ctx, item)
	case "metric":
		return processMetric(ctx, item)
	default:
		return entities.Result{
			ItemID: item.ID,
			Error:  fmt.Sprintf("unknown item type: %q", item.Type),
		}
	}
}

// processText upper-cases all string values in the payload.
func processText(_ context.Context, item entities.Item) entities.Result {
	out := make(map[string]any, len(item.Payload))
	for k, v := range item.Payload {
		if s, ok := v.(string); ok {
			out[k] = strings.ToUpper(s)
		} else {
			out[k] = v
		}
	}
	return entities.Result{ItemID: item.ID, Output: out}
}

// processMetric doubles all numeric values in the payload.
func processMetric(_ context.Context, item entities.Item) entities.Result {
	out := make(map[string]any, len(item.Payload))
	for k, v := range item.Payload {
		if n, ok := v.(float64); ok {
			out[k] = n * 2
		} else {
			out[k] = v
		}
	}
	return entities.Result{ItemID: item.ID, Output: out}
}
