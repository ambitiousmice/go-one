package config

import (
	"go-one/common/context"
	"go-one/common/log"
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
	Name       string   `yaml:"name"`
	GroupNames []string `yaml:"groupNames"`
}
