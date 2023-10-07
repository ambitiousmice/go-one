package context

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

const (
	contextFile = "context.yaml"
)

var yamlFile = contextFile
var oneConfig OneConfig

func SetYamlFile(yaml string) {
	yamlFile = yaml
}

func GetOneConfig() OneConfig {
	return oneConfig
}

func InitConfig() error {
	yamlFileBytes, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		return err
	} // 将读取的yaml文件解析为响应的 struct
	err = yaml.Unmarshal(yamlFileBytes, &oneConfig)
	if err != nil {
		return err
	}
	return nil
}
