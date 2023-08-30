package game

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

const (
	YamlFile = "game.yaml"
)

var yamlFile = YamlFile
var gameConfig GameConfig

func SetYamlFile(yaml string) {
	yamlFile = yaml
}

type GameConfig struct {
	Server ServerConfig

	Params map[string]interface{}
}

func InitGameConfig() error {
	yamlFile, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		fmt.Println(err.Error())
	} // 将读取的yaml文件解析为响应的 struct
	err = yaml.Unmarshal(yamlFile, &gameConfig)
	if err != nil {
		return err
	}
	return nil
}

type ServerConfig struct {
	ListenAddr             string `yaml:"listenAddr"`
	GoMaxProcs             int    `yaml:"goMaxProcs"`
	HeartbeatCheckInterval int    `yaml:"heartbeatCheckInterval"`
	GateTimeout            int32  `yaml:"gateTimeout"`
}
