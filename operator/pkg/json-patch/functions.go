package json_patch

import (
	"encoding/json"
)

type Document []Patch

func (pd *Document) Add(op, path string, value any) *Document {
	*pd = append(*pd, Patch{op, path, value})
	return pd
}

func (pd *Document) Json() ([]byte, error) {
	return json.Marshal(pd)
}

type Patch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}
