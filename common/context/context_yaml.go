package context

import (
	"fmt"
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

func ReadYaml() error {
	yamlFile, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		fmt.Println(err.Error())
	} // 将读取的yaml文件解析为响应的 struct
	err = yaml.Unmarshal(yamlFile, &oneConfig)
	if err != nil {
		return err
	}
	return nil
}
