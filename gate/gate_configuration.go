package gate

import (
	"fmt"
	"go-one/common/entity"
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

func InitGateConfig() error {
	yamlFile, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		fmt.Println(err.Error())
	} // 将读取的yaml文件解析为响应的 struct
	err = yaml.Unmarshal(yamlFile, &gateConfig)
	if err != nil {
		panic(err.Error())
	}
	return nil
}

type ServerConfig struct {
	ListenAddr             string `yaml:"listenAddr"`
	GoMaxProcs             int    `yaml:"goMaxProcs"`
	HeartbeatCheckInterval int    `yaml:"heartbeatCheckInterval"`
	ClientTimeout          int32  `yaml:"clientTimeout"`
	NeedLogin              bool   `yaml:"needLogin"`
}
