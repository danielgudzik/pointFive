package pipeline

import (
	"github.com/example/pointfive/entities"
	wp "github.com/example/pointfive/utils/workerpool"
)

// ItemPipeline processes entities.Item → entities.Result jobs.
type ItemPipeline = wp.Pipeline[entities.Item, entities.Result]

// NewItemPipeline creates an ItemPipeline wired to processItem.
func NewItemPipeline(cfg entities.PipelineSettings) *ItemPipeline {
	return wp.New[entities.Item, entities.Result](cfg.WorkerCount, cfg.Log, processItem)
}
