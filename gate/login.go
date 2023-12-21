package gate

import (
	"errors"
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/common/consts"
)

type LoginManager interface {
	TokenLogin(token string) (*LoginResult, error)
}

type LoginResult struct {
	EntityID int64
}

func Login(manager LoginManager, param common_proto.LoginReq) (*LoginResult, error) {
	switch param.AccountType {
	case consts.TokenLogin:
		return manager.TokenLogin(param.Account)
	default:
		return nil, errors.New("unknown login type")
	}
}
