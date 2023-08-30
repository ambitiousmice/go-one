package pktconn

import (
	"bytes"

	"github.com/vmihailenco/msgpack"
)

// BytesPackMsgPacker packs and unpacks message in MessagePack format
type BytesPackMsgPacker struct{}

// PackMsg packs message to bytes in MessagePack format
func (mp BytesPackMsgPacker) PackMsg(msg interface{}, buf []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(buf)

	encoder := msgpack.NewEncoder(buffer)
	err := encoder.Encode(msg)
	if err != nil {
		return buf, err
	}
	buf = buffer.Bytes()
	return buf, nil
}

// UnpackMsg unpacks bytes in MessagePack format to message
func (mp BytesPackMsgPacker) UnpackMsg(data []byte, msg interface{}) error {
	err := msgpack.Unmarshal(data, msg)
	return err
}
