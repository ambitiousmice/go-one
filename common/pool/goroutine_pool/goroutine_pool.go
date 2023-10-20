package goroutine_pool

import (
	"github.com/panjf2000/ants/v2"
	"go-one/common/log"
)

var goroutinePool *ants.Pool

func IsEnable() bool {
	if goroutinePool == nil {
		return false
	}
	return true
}
func Init(poolSize int) {
	if poolSize <= 0 {
		return
	}
	pool, err := ants.NewPool(poolSize)
	if err != nil {
		log.Panicf("init goroutine pool error:%s", err.Error())
	}
	goroutinePool = pool
}

func Submit(task func()) error {
	return goroutinePool.Submit(task)
}

func Release() {
	goroutinePool.Release()
}

func Running() int {
	return goroutinePool.Running()
}
