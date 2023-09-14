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
	Server             ServerConfig           `yaml:"server"`
	RoomManagerConfigs []RoomManagerConfig    `yaml:"room-manager-configs"`
	Params             map[string]interface{} `yaml:"params"`
}

func InitGameConfig() {
	yamlFile, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		fmt.Println(err.Error())
	} // 将读取的yaml文件解析为响应的 struct
	err = yaml.Unmarshal(yamlFile, &gameConfig)
	if err != nil {
		panic("init game config error: " + err.Error())
	}

}

type ServerConfig struct {
	ListenAddr             string `yaml:"listenAddr"`
	GoMaxProcs             int    `yaml:"goMaxProcs"`
	HeartbeatCheckInterval int    `yaml:"heartbeatCheckInterval"`
	GateTimeout            int32  `yaml:"gateTimeout"`
}

type RoomManagerConfig struct {
	RoomType         string `yaml:"room-type"`
	RoomMaxPlayerNum int    `yaml:"room-max-player-num"`
	RoomIDStart      int64  `yaml:"room-id-start"`
	RoomIDEnd        int64  `yaml:"room-id-end"`
	MatchStrategy    string `yaml:"match-strategy"`
}
