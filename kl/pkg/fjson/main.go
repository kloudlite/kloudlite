package fjson

import (
	"bytes"
	"encoding/json"

	"github.com/kloudlite/kl/pkg/functions"
)

func Marshal(obj any) ([]byte, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, functions.NewE(err)
	}

	bjson := new(bytes.Buffer)
	if err = json.Indent(bjson, b, "", "  "); err != nil {
		return nil, functions.NewE(err)
	}

	return bjson.Bytes(), nil
}

func Unmarshal(data []byte, obj any) error {
	return json.Unmarshal(data, obj)
}
