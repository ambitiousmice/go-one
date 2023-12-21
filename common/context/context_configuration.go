package context

import (
	"flag"
	"github.com/ambitiousmice/go-one/common/log"
	"github.com/ambitiousmice/go-one/common/register"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func init() {
	flag.StringVar(&yamlFile, "cc", "context.yaml", "set config file path")
}

var yamlFile string
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

func InitConfigFromNacos() {
	var DataId string
	if len(oneConfig.Nacos.DataID) == 0 {
		DataId = oneConfig.Nacos.Instance.Service + "-" + oneConfig.Nacos.Instance.ClusterName + ".yaml"
	} else {
		DataId = oneConfig.Nacos.DataID
	}
	content, err := register.ConfigClient.GetConfig(vo.ConfigParam{
		DataId: DataId,
		Group:  oneConfig.Nacos.Instance.GroupName,
	})

	if err != nil {
		log.Panic(err)
	}

	err = yaml.Unmarshal([]byte(content), &oneConfig)

	if err != nil {
		log.Panic(err)
	}

	configFromNacos = content

}
