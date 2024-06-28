package egob

import (
	"bytes"
	"encoding/gob"

	"github.com/kloudlite/kl/pkg/functions"
)

func Marshal(obj any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(obj); err != nil {
		return nil, functions.NewE(err)
	}
	return buf.Bytes(), nil
}

func Unmarshal(data []byte, obj any) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(obj); err != nil {
		return functions.NewE(err)
	}
	return nil
}
