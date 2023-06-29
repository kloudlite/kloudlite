package types

import (
	"encoding/base64"
	fn "kloudlite.io/pkg/functions"
)

type M map[string]interface{}

type Pagination struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
}

type CursorSortBy struct {
	Field     string        `json:"field"`
	Direction SortDirection `json:"sortDirection"`
}

type Cursor string

func CursorToBase64(c Cursor) string {
	return base64.StdEncoding.EncodeToString([]byte(c))
}

func CursorFromBase64(b string) (Cursor, error) {
	b2, err := base64.StdEncoding.DecodeString(b)
	if err != nil {
		return Cursor(""), err
	}
	return Cursor(b2), nil
}

type CursorPagination struct {
	First *int64  `json:"first"`
	After *string `json:"after,omitempty"`

	Last   *int64  `json:"last,omitempty"`
	Before *string `json:"before,omitempty"`

	OrderBy       string        `json:"orderBy,omitempty"`
	SortDirection SortDirection `json:"sortDirection,omitempty" graphql:"enum=ASC;DESC"`
}

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

var DefaultCursorPagination = CursorPagination{
	First:         fn.New(int64(10)),
	After:         nil,
	OrderBy:       "_id",
	SortDirection: SortDirectionAsc,
}
