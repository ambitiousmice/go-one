package game_client

import (
	"fmt"
	"github.com/ambitiousmice/go-one/common/context"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

const (
	YamlFile = "context_client.yaml"
)

var yamlFile = YamlFile
var Config config

func SetYamlFile(yaml string) {
	yamlFile = yaml
}

type config struct {
	ServerConfig      serverConfig              `yaml:"server"`
	IDGeneratorConfig context.IDGeneratorConfig `yaml:"id_generator"`
}

type serverConfig struct {
	UseLoadBalancer bool   `yaml:"use-load-balancer"`
	LoadBalancerUrl string `yaml:"load-balancer-url"`
	Partition       int32  `yaml:"partition"`
	Kcp             bool   `yaml:"kcp"`
	Websocket       bool   `yaml:"websocket"`
	ServerHost      string `yaml:"server-host"`
	ClientNum       int    `yaml:"client_num"`
	ApiUrl          string `yaml:"api-url"`
	ApiToken        string `yaml:"api-token"`
	DataPath        string `yaml:"data-path"`
}

func InitConfig() error {
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
