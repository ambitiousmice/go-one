package context

import (
	"go-one/common/register"
)

func Init() error {
	err := ReadYaml()
	if err != nil {
		return nil
	}

	register.Run(oneConfig.Nacos)

	InitIDGenerator(oneConfig.IDGeneratorConfig)

	return nil
}
