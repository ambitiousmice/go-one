package processor_center

import (
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/log"
)

var dispatcherPlayerProcessors []chan ProcessorTask

func init() {
	poolSize := context.GetOneConfig().PoolConfig.GoroutinePoolSize
	if poolSize <= 2 {
		poolSize = 10
	}
	dispatcherPlayerProcessors = make([]chan ProcessorTask, poolSize)
	for i := 0; i < poolSize; i++ {
		dispatcherPlayerProcessors[i] = make(chan ProcessorTask, consts.GameServiceProcessorQueueSize)
		tempI := i
		go func() {
			for {
				select {
				case task := <-dispatcherPlayerProcessors[tempI]:
					func() {
						defer func() {
							if r := recover(); r != nil {
								log.Errorf("handle player business,Recover from panic: %v", r)
							}
						}()
						task()
					}()
				default:
					continue
				}
			}
		}()
	}
}

type ProcessorTask func()

func SubmitProcessorTask(entityID int64, task ProcessorTask) {
	index := entityID % int64(len(dispatcherPlayerProcessors))
	dispatcherPlayerProcessors[index] <- task
}
