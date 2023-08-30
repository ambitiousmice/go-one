package pktconn

var (
	// MSG_PACKER is used for packing and unpacking network data
	MSG_PACKER MsgPacker = BytesPackMsgPacker{}
)

// MsgPacker is used to packs and unpacks messages
type MsgPacker interface {
	PackMsg(msg interface{}, buf []byte) ([]byte, error)
	UnpackMsg(data []byte, msg interface{}) error
}
