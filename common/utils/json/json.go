package json

import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func ToBytes(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func ToString(v interface{}) (string, error) {
	return json.MarshalToString(v)
}

func ParseBytes(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func ParseString(data string, v interface{}) error {
	return json.UnmarshalFromString(data, v)
}
