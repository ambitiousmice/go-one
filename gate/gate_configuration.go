package gate

import (
	"go-one/common/entity"
	"go-one/common/log"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

const (
	YamlFile = "gate.yaml"
)

var yamlFile = YamlFile
var gateConfig GateConfig

func SetYamlFile(yaml string) {
	yamlFile = yaml
}

func GetGateConfig() GateConfig {
	return gateConfig
}

type GateConfig struct {
	Server                ServerConfig
	GameDispatcherConfigs []entity.GameDispatcherConfig `yaml:"game-dispatcher"`
	Params                map[string]interface{}
}

func InitConfig() {
	yamlFileBytes, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		log.Panic(err.Error())
	} // 将读取的yaml文件解析为响应的 struct
	err = yaml.Unmarshal(yamlFileBytes, &gateConfig)
	if err != nil {
		log.Panic(err.Error())
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
