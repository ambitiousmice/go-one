package main

import (
	"github.com/ambitiousmice/go-one/common/common_proto"
	"github.com/ambitiousmice/go-one/gate"
)

type DemoLoginManager struct {
	loginServerUrl string
}

func NewDemoLoginManager(loginServerUrl string) *DemoLoginManager {
	return &DemoLoginManager{
		loginServerUrl: loginServerUrl,
	}
}

type validateTokenReq struct {
	AccessToken string `json:"accessToken"`
}

type resultResp struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type ValidateTokenResp struct {
	Id        string `json:"id"`
	NickName  string `json:"nickName"`
	AvatarUrl string `json:"avatarUrl"`
	Sex       int32  `json:"sex"`
	RoleId    int    `json:"roleId"`
}

func (manager *DemoLoginManager) Login(param *common_proto.LoginReq) (*gate.LoginResult, error) {
	/*data := &ValidateTokenResp{}
	result := &resultResp{Data: data}

	err := utils.Post(manager.loginServerUrl, validateTokenReq{AccessToken: token}, &result)

	if err != nil {
		return nil, err
	}

	if result.Code != "0" {
		return nil, errors.New(result.Msg)
	}

	EntityID, err := strconv.ParseInt(data.Id, 10, 64)

	if err != nil {
		return nil, err
	}

	return &gate.LoginResult{
		EntityID: EntityID,
	}, nil*/

	return &gate.LoginResult{
		Success: true,
	}, nil
}
