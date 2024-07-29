package pool

import (
	"github.com/ambitiousmice/go-one/common/pool/fixed_channel_pool"
	"github.com/ambitiousmice/go-one/common/pool/goroutine_pool"
)

func InitPool(config Config) {
	if config.GoroutinePoolEnable {
		goroutine_pool.Init(config.GoroutinePoolSize)
	}

	if config.FixedChannelPoolEnable {
		fixed_channel_pool.Init(config.FixedChannelPoolSize)
	}
}
