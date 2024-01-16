package pktconn

import "google.golang.org/protobuf/proto"

// PbMsgPacker packs and unpacks message in MessagePack format
type PbMsgPacker struct{}

// PackMsg packs message to bytes in MessagePack format
func (mp PbMsgPacker) PackMsg(msg interface{}, buf []byte) ([]byte, error) {
	buf, err := proto.Marshal(msg.(proto.Message))
	if err != nil {
		return buf, err
	}
	return buf, nil
}

// UnpackMsg unpacks bytes in MessagePack format to message
func (mp PbMsgPacker) UnpackMsg(data []byte, msg interface{}) error {
	err := proto.Unmarshal(data, msg.(proto.Message))
	return err
}
