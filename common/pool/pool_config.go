package pool

type Config struct {
	GoroutinePoolEnable bool `yaml:"goroutine-pool-enable"`
	GoroutinePoolSize   int  `yaml:"goroutine-pool-size"`

	FixedChannelPoolEnable bool `yaml:"fixed-channel-pool-enable"`
	FixedChannelPoolSize   int  `yaml:"fixed-channel-pool-size"`
}
