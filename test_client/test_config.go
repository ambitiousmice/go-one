package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

const (
	YamlFile = "context_test.yaml"
)

var yamlFile = YamlFile
var Config config

func SetYamlFile(yaml string) {
	yamlFile = yaml
}

type config struct {
	ServerConfig serverConfig `yaml:"server"`
}

type serverConfig struct {
	Kcp       bool   `yaml:"kcp"`
	Websocket bool   `yaml:"websocket"`
	IP        string `yaml:"ip"`
	Port      string `yaml:"port"`
	ClientNum int    `yaml:"client_num"`
}

func ReadTestConfig() error {
	yamlFileBytes, err := ioutil.ReadFile(yamlFile)
	if err != nil {
		fmt.Println(err.Error())
	} // 将读取的yaml文件解析为响应的 struct
	err = yaml.Unmarshal(yamlFileBytes, &Config)
	if err != nil {
		return err
	}
	return nil
}
