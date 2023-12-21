package common

import (
	"github.com/ambitiousmice/go-one/common/pktconn"
)

func UnPackMsg(param []byte, obj any) error {
	err := pktconn.MSG_PACKER.UnpackMsg(param, obj)
	return err
}
