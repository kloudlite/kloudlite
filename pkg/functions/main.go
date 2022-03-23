package functions

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
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
	if e != nil {
		return "", e
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
