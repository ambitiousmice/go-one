package pktconn

var (
	// MSG_PACKER is used for packing and unpacking network data
	//MSG_PACKER MsgPacker = BytesPackMsgPacker{}
	//MSG_PACKER MsgPacker = JSONMsgPacker{}
	MSG_PACKER MsgPacker = PbMsgPacker{}
)

// MsgPacker is used to packs and unpacks messages
type MsgPacker interface {
	PackMsg(msg interface{}, buf []byte) ([]byte, error)
	UnpackMsg(data []byte, msg interface{}) error
}

func SetMsgPacker(msgPacker MsgPacker) {
	MSG_PACKER = msgPacker
}
