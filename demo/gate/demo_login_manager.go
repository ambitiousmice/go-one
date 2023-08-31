package main

import (
	"errors"
	"go-one/common/utils"
	"go-one/gate"
	"strconv"
)

type EliteStarLoginManager struct {
	loginServerUrl string
}

func NewEliteStarLoginManager(loginServerUrl string) *EliteStarLoginManager {
	return &EliteStarLoginManager{
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

func (manager *EliteStarLoginManager) TokenLogin(token string) (*gate.LoginResult, error) {
	data := &ValidateTokenResp{}
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
	}, nil
}
