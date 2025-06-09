package types

type M map[string]interface{}

type Pagination struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
}
