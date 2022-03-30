package functions

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
)

func ToBytes(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ToBase64String(v interface{}) (string, error) {
	b, e := ToBytes(v)
	return base64.StdEncoding.EncodeToString(b), e
}

func ToBase64StringFromJson(v interface{}) (string, error) {
	b, e := json.Marshal(v)
	return base64.StdEncoding.EncodeToString(b), e
}
