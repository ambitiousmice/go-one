package entity

type GameDispatcherConfig struct {
	Game         string `yaml:"game"`
	ChannelNum   uint8  `yaml:"channel-num"`
	LoadBalancer string `yaml:"load-balancer"`
}
