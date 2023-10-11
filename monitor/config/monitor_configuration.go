package config

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func init() {
	flag.StringVar(&yamlFile, "gc", "context.yaml", "set config file path")
}

var yamlFile string
var config Config

func SetYamlFile(yaml string) {
	yamlFile = yaml
}

type Config struct {
	Gate Gate `yaml:"gate"`
}

func InitConfig() {
	yamlFileBytes, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		fmt.Println(err.Error())
	} // 将读取的yaml文件解析为响应的 struct
	err = yaml.Unmarshal(yamlFileBytes, &config)
	if err != nil {
		panic("init monitor config error: " + err.Error())
	}
}

type Gate struct {
	Name string `yaml:"name"`
}
