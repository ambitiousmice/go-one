package context

import (
	"go-one/common/log"
	"go-one/common/register"
)

type OneConfig struct {
	Nacos             register.NacosConf `yaml:"nacos"`
	Logger            log.Config         `yaml:"logger"`
	IDGeneratorConfig IDGeneratorConfig  `yaml:"id_generator"`
}
