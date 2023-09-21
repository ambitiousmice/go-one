package game

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

const (
	YamlFile = "Game.yaml"
)

var yamlFile = YamlFile
var gameConfig GameConfig

func SetYamlFile(yaml string) {
	yamlFile = yaml
}

type GameConfig struct {
	Server              ServerConfig           `yaml:"server"`
	SceneManagerConfigs []SceneManagerConfig   `yaml:"scene-manager-configs"`
	Params              map[string]interface{} `yaml:"params"`
}

func InitGameConfig() {
	yamlFile, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		fmt.Println(err.Error())
	} // 将读取的yaml文件解析为响应的 struct
	err = yaml.Unmarshal(yamlFile, &gameConfig)
	if err != nil {
		panic("init Game config error: " + err.Error())
	}

}

type ServerConfig struct {
	Game                   string `yaml:"Game"`
	ListenAddr             string `yaml:"listenAddr"`
	GoMaxProcs             int    `yaml:"goMaxProcs"`
	HeartbeatCheckInterval int    `yaml:"heartbeatCheckInterval"`
	GateTimeout            int32  `yaml:"gateTimeout"`
}

type SceneManagerConfig struct {
	SceneType         string `yaml:"scene-type"`
	SceneMaxPlayerNum int    `yaml:"scene-max-player-num"`
	SceneIDStart      int64  `yaml:"scene-id-start"`
	SceneIDEnd        int64  `yaml:"scene-id-end"`
	MatchStrategy     string `yaml:"match-strategy"`
}
