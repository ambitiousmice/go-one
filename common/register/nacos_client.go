package register

import (
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"go-one/common/consts"
	"go-one/common/log"
	"strconv"
	"strings"
)

var NacosClient naming_client.INamingClient

type NacosConf struct {
	Host      string   `yaml:"host"`
	Instance  Instance `yaml:"instance"`
	Namespace string   `yaml:"namespace"`
	LogLevel  string   `yaml:"logLevel"`
}

type Instance struct {
	Ip          string            `yaml:"ip"`
	Port        uint64            `yaml:"port"`
	Service     string            `yaml:"service"`
	Metadata    map[string]string `yaml:"metadata"`
	GroupName   string            `yaml:"groupName"`
	ClusterName string            `yaml:"clusterName"`
}

func Run(config NacosConf) {
	if len(config.Host) == 0 {
		panic("nacos server addr is empty")
	}

	serverAddresses := strings.Split(config.Host, ",")
	sc := make([]constant.ServerConfig, len(serverAddresses))

	for i, addr := range serverAddresses {
		serverAddress := strings.Split(addr, ":")
		host := serverAddress[0]
		port, err := strconv.ParseUint(serverAddress[1], 10, 64)
		if err != nil {
			panic(err)
		}

		sc[i] = *constant.NewServerConfig(host, port, constant.WithContextPath("/nacos"))
	}

	if len(config.LogLevel) == 0 {
		config.LogLevel = "debug"
	}

	//create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithNamespaceId(config.Namespace),
		constant.WithUpdateCacheWhenEmpty(true),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir("/tmp/nacos/log"),
		constant.WithCacheDir("/tmp/nacos/cache"),
		constant.WithLogLevel(config.LogLevel),
	)

	// create naming client
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)

	if err != nil {
		panic(err)
	}

	instance := config.Instance

	if len(instance.GroupName) != 0 {
		instance.Metadata[consts.Partition] = instance.GroupName
		instance.Metadata[consts.ClusterId] = instance.ClusterName
	}
	//Register
	registerServiceInstance(client, vo.RegisterInstanceParam{
		Ip:          instance.Ip,
		Port:        instance.Port,
		ServiceName: instance.Service,
		GroupName:   instance.GroupName,
		ClusterName: instance.ClusterName,
		Weight:      10,
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    instance.Metadata,
	})
	NacosClient = client
}

func registerServiceInstance(client naming_client.INamingClient, param vo.RegisterInstanceParam) {
	success, err := client.RegisterInstance(param)
	if !success || err != nil {
		panic("RegisterServiceInstance failed!" + err.Error())
	}
	log.Infof("RegisterServiceInstance,param:%+v,result:%+v \n\n", param, success)
}
