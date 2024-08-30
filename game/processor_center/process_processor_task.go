package processor_center

import (
	"github.com/ambitiousmice/go-one/common/pool/fixed_channel_pool"
)

type ProcessorTask func()

// deprecated,  please use fixed_channel_pool replace
func SubmitProcessorTask(entityID int64, task ProcessorTask) {
	fixed_channel_pool.Submit(entityID, task)
}
