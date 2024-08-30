package fixed_channel_pool

import (
	"github.com/ambitiousmice/go-one/common/log"
	"hash/fnv"
	"runtime"
)

var taskChannels []chan func()
var fixedChannelPoolSize int64

func Init(poolSize int) {
	if poolSize <= 0 {
		poolSize = runtime.NumCPU()
	}
	fixedChannelPoolSize = int64(poolSize)
	taskChannels = make([]chan func(), poolSize)
	for i := 0; i < poolSize; i++ {
		taskChannels[i] = make(chan func(), 102400)
		tempI := i
		go func() {
			log.Infof("fixed pool channels[%d] start", tempI)
			for {
				select {
				case task := <-taskChannels[tempI]:
					func() {
						defer func() {
							if r := recover(); r != nil {
								log.Errorf("handle task error,recover from panic: %v", r)
							}
						}()
						task()
					}()
				}
			}
		}()
	}

	log.Infof("fixed channel pool init success,pool size:%d", poolSize)
}

func Submit(fixedID int64, task func()) {
	index := fixedID % fixedChannelPoolSize
	taskChannels[index] <- task
}

func SubmitByStr(fixedID string, task func()) {
	index := hash(fixedID) % fixedChannelPoolSize
	taskChannels[index] <- task
}

func hash(s string) int64 {
	h := fnv.New64a()       // 创建一个新的 FNV-1a 64-bit 哈希对象
	h.Write([]byte(s))      // 将字符串转换为字节切片并写入哈希对象
	return int64(h.Sum64()) // 返回 64-bit 的哈希值
}
