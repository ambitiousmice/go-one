package game_client

import (
	"go-one/common/log"
	"go-one/common/pktconn"
)

func UnPackMsg(param []byte, obj any) {
	err := pktconn.MSG_PACKER.UnpackMsg(param, obj)
	if err != nil {
		log.Errorf("unpack msg error: %s", err.Error())
	}
}

func PackMsg(obj any) []byte {
	bytes, err := pktconn.MSG_PACKER.PackMsg(obj, nil)
	if err != nil {
		log.Errorf("unpack msg error: %s", err.Error())
	}

	return bytes
}
