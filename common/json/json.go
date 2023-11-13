package json

import jsoniter "github.com/json-iterator/go"

var jsonI = jsoniter.ConfigCompatibleWithStandardLibrary

func UnmarshalFromString(str string, v interface{}) error {
	return jsonI.UnmarshalFromString(str, v)
}

func Unmarshal(data []byte, v interface{}) error {
	return jsonI.Unmarshal(data, v)
}

func Marshal(v interface{}) ([]byte, error) {
	return jsonI.Marshal(v)
}

func MarshalToString(v interface{}) (string, error) {
	return jsonI.MarshalToString(v)
}
