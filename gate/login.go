package gate

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
)

type LoginManager interface {
	Login(param *common_proto.LoginReq) (*LoginResult, error)
}

type LoginResult struct {
	Success  bool
	Region   int32
	EntityID int64
}

func Login(manager LoginManager, param *common_proto.LoginReq) (*LoginResult, error) {
	return manager.Login(param)
}
