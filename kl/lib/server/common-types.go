package server

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Status struct {
	IsReady bool `json:"isReady"`
	Message struct {
		RawMessage string `json:"RawMessage"`
	} `json:"message"`
}

type Edges[T any] []Edge[T]

type Edge[T any] struct {
	Node T
}
