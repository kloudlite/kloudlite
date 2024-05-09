package server

import "encoding/json"

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Status struct {
	IsReady bool `json:"isReady"`
	Message struct {
		RawMessage json.RawMessage `json:",inline"`
	} `json:"message"`
}

type Edges[T any] []Edge[T]

type Edge[T any] struct {
	Node T
}
