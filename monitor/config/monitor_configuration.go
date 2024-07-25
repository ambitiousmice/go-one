package config

import (
	"github.com/ambitiousmice/go-one/common/context"
	"github.com/ambitiousmice/go-one/common/log"
	"gopkg.in/yaml.v3"
)

var config Config

func GetConfig() Config {
	return config
}

type Config struct {
	Gate Gate `yaml:"gate"`
}

func InitConfig() {
	configFromNacos := context.GetConfigFromNacos()
	configByte := []byte(configFromNacos)
	err := yaml.Unmarshal(configByte, &config)
	if err != nil {
		log.Panic(err.Error())
	}
}

type Gate struct {
	Name                      string              `yaml:"name"`
	CommunicationTimeoutValue int64               `yaml:"communicationTimeoutValue"`
	GroupInfos                map[string][]string `yaml:"groupInfos"` // groupID -> [regionID1, regionID2]
}
