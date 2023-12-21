package context

import (
	"errors"
	"github.com/ambitiousmice/go-one/common/consts"
	"github.com/ambitiousmice/go-one/common/idgenerator"
	"github.com/ambitiousmice/go-one/common/log"
)

var idGenerator idgenerator.IDGenerator

type IDGeneratorConfig struct {
	Type   string `yaml:"type"`
	NodeID int64  `yaml:"node_id"`
}

func InitIDGenerator(config IDGeneratorConfig) error {
	if config.Type == "" {
		return nil
	}

	switch config.Type {

	case consts.Snowflake:
		node, err := idgenerator.NewNode(config.NodeID)
		if err != nil {
			return err
		}
		idGenerator = node
		log.Info("init snowflake id generator success")
	default:
		return errors.New("unknown id generator type")
	}
	return nil
}

func NextClientID() string {
	return idGenerator.NextIDStr()
}

func NextEntityID() int64 {
	return idGenerator.NextID()
}
