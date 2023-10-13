package gate

import (
	"go-one/common/context"
	"go-one/common/entity"
	"go-one/common/log"
	"gopkg.in/yaml.v3"
)

var gateConfig GateConfig

func GetGateConfig() GateConfig {
	return gateConfig
}

type GateConfig struct {
	Server                ServerConfig
	GameDispatcherConfigs []entity.GameDispatcherConfig `yaml:"game-dispatcher"`
	Params                map[string]interface{}
}

func InitConfig(localConfigs ...any) {
	configFromNacos := context.GetConfigFromNacos()
	configFromNacosBytes := []byte(configFromNacos)
	err := yaml.Unmarshal(configFromNacosBytes, &gateConfig)
	if err != nil {
		log.Panic(err.Error())
	}

	if len(localConfigs) == 0 {
		return
	}

	for _, config := range localConfigs {
		err = yaml.Unmarshal(configFromNacosBytes, config)
		if err != nil {
			log.Panic(err.Error())
		}
	}
}

type ServerConfig struct {
	ListenAddr                       string `yaml:"listenAddr"`
	WebsocketListenAddr              string `yaml:"websocketListenAddr"`
	GoMaxProcs                       int    `yaml:"goMaxProcs"`
	HeartbeatCheckInterval           int    `yaml:"heartbeatCheckInterval"`
	ClientTimeout                    int32  `yaml:"clientTimeout"`
	NeedLogin                        bool   `yaml:"needLogin"`
	DispatcherClientPacketQueuesSize int    `yaml:"dispatcherClientPacketQueuesSize"`
}
