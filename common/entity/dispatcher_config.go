package entity

type GameDispatcherConfig struct {
	Game         string `yaml:"game"`
	GroupName    string `yaml:"group-name"`
	ChannelNum   uint8  `yaml:"channel-num"`
	LoadBalancer string `yaml:"load-balancer"`
}
