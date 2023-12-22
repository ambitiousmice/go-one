package gate

import (
	"errors"
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
)

type LoginManager interface {
	Login(param *common_proto.LoginReq) (*LoginResult, error)
}

type LoginResult struct {
	Success  bool
	EntityID int64
}

func Login(manager LoginManager, param *common_proto.LoginReq) (*LoginResult, error) {
	switch param.AccountType {
	case consts.TokenLogin:
		return manager.Login(param)
	default:
		return nil, errors.New("unknown login type")
	}
}
