package gate

import (
	"errors"
	"go-one/common/common_proto"
	"go-one/common/consts"
)

type LoginManager interface {
	TokenLogin(token string) (*LoginResult, error)
}

type LoginResult struct {
	EntityID int64
}

func Login(manager LoginManager, param common_proto.EnterGameReq) (*LoginResult, error) {
	switch param.AccountType {
	case consts.TokenLogin:
		return manager.TokenLogin(param.Account)
	default:
		return nil, errors.New("unknown login type")
	}
}
