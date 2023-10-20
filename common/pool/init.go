package pool

import "go-one/common/pool/goroutine_pool"

func InitPool(config Config) {
	goroutine_pool.Init(config.GoroutinePoolSize)
}
