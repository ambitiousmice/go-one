package context

import (
	"go-one/common/consts"
	"go-one/common/idgenerator"
	"go-one/common/log"
)

var idGenerator idgenerator.IDGenerator

type IDGeneratorConfig struct {
	Type   string `yaml:"type"`
	NodeID int64  `yaml:"node_id"`
}

func InitIDGenerator(config IDGeneratorConfig) {
	if config.Type == "" {
		return
	}

	switch config.Type {

	case consts.Snowflake:
		node, err := idgenerator.NewNode(config.NodeID)
		if err != nil {
			panic(err)
		}
		idGenerator = node
		log.Info("init snowflake id generator success")
	default:
		panic("unknown id generator type")
	}
}

func NextClientID() string {
	return idGenerator.NextID()
}
