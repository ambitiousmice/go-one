package game

import (
	"go-one/common/context"
	"go-one/common/log"
	"gopkg.in/yaml.v3"
)

var gameConfig GameConfig
var LocalConfig any

func SetLocalConfig(c any) {
	LocalConfig = c
}

type GameConfig struct {
	Server              ServerConfig           `yaml:"server"`
	SceneManagerConfigs []SceneManagerConfig   `yaml:"scene-manager-configs"`
	Params              map[string]interface{} `yaml:"params"`
}

func InitConfig(localConfigs ...any) {
	configFromNacos := context.GetConfigFromNacos()
	configByte := []byte(configFromNacos)
	err := yaml.Unmarshal(configByte, &gameConfig)
	if err != nil {
		log.Panic(err.Error())
	}

	if len(localConfigs) == 0 {
		return
	}

	for _, config := range localConfigs {
		err = yaml.Unmarshal(configByte, config)
		if err != nil {
			log.Panic(err.Error())
		}
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
	EnableAOI         bool    `yaml:"enable-aoi"`
	AOIDistance       float32 `yaml:"aoi-distance"`
	SceneType         string  `yaml:"scene-type"`
	SceneMaxPlayerNum int     `yaml:"scene-max-player-num"`
	SceneIDStart      int64   `yaml:"scene-id-start"`
	SceneIDEnd        int64   `yaml:"scene-id-end"`
	MatchStrategy     string  `yaml:"match-strategy"`
}
