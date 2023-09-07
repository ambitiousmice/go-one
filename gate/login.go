package gate

import (
	"errors"
	"go-one/common/consts"
	"go-one/common/proto"
)

type LoginManager interface {
	TokenLogin(token string) (*LoginResult, error)
}

type LoginResult struct {
	EntityID int64
}

func Login(manager LoginManager, param proto.EnterGameReq) (*LoginResult, error) {
	switch param.AccountType {
	case consts.TokenLogin:
		return manager.TokenLogin(param.Account)
	default:
		return nil, errors.New("unknown login type")
	}
}
