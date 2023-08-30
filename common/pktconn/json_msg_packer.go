package pktconn

import (
	"bytes"
	"encoding/json"
)

type JSONMsgPacker struct{}

// PackMsg packs message to bytes of JSON format
func (mp JSONMsgPacker) PackMsg(msg interface{}, buf []byte) ([]byte, error) {
	buffer := bytes.NewBuffer(buf)
	jsonEncoder := json.NewEncoder(buffer)
	err := jsonEncoder.Encode(msg)
	if err != nil {
		return buf, err
	}

	buf = buffer.Bytes()

	return buf[:len(buf)-1], nil
}

// UnpackMsg unpacks bytes of JSON format to message
func (mp JSONMsgPacker) UnpackMsg(data []byte, msg interface{}) error {
	err := json.Unmarshal(data, msg)
	return err
}
