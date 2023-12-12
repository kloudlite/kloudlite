package repos

import (
	"context"
	"encoding/base64"
	"github.com/kloudlite/api/pkg/functions"
	"time"
)

type Entity interface {
	GetPrimitiveID() ID
	GetId() ID
	SetId(id ID)
	GetCreationTime() time.Time
	GetUpdateTime() time.Time
	SetCreationTime(time.Time)
	SetUpdateTime(time.Time)
	IsZero() bool

	IncrementRecordVersion()
	GetRecordVersion() int
	IsMarkedForDeletion() bool
}

type Opts map[string]interface{}
type SortOpts map[string]int32
type Filter map[string]interface{}
type Query struct {
	Filter Filter
	Sort   map[string]interface{}
}

type MatchType string

const (
	MatchTypeExact = "exact"
	MatchTypeArray = "array"
	MatchTypeRegex = "regex"
)

type MatchFilter struct {
	MatchType MatchType `json:"matchType" graphql:"enum=exact;array;regex;"`
	Exact     any       `json:"exact,omitempty"`
	Array     []any     `json:"array,omitempty"`
	Regex     *string   `json:"regex,omitempty"`
}

type ID string

type PageInfo struct {
	StartCursor string
	EndCursor   string
	HasNextPage *bool
	HasPrevPage *bool
}

type RecordEdge[T Entity] struct {
	Node   T
	Cursor string
}

type PaginatedRecord[T Entity] struct {
	Edges      []RecordEdge[T]
	PageInfo   PageInfo
	TotalCount int64
}

type UpdateOpts struct {
	Upsert bool
}

type DbRepo[T Entity] interface {
	NewId() ID
	Find(ctx context.Context, query Query) ([]T, error)
	FindOne(ctx context.Context, filter Filter) (T, error)
	FindPaginated(ctx context.Context, filter Filter, pagination CursorPagination) (*PaginatedRecord[T], error)
	FindById(ctx context.Context, id ID) (T, error)
	Create(ctx context.Context, data T) (T, error)
	Exists(ctx context.Context, filter Filter) (bool, error)

	Count(ctx context.Context, filter Filter) (int64, error)

	// upsert
	Upsert(ctx context.Context, filter Filter, data T) (T, error)
	UpdateMany(ctx context.Context, filter Filter, updatedData map[string]any) error
	UpdateById(ctx context.Context, id ID, updatedData T, opts ...UpdateOpts) (T, error)
	UpdateOne(ctx context.Context, filter Filter, updatedData T, opts ...UpdateOpts) (T, error)
	SilentUpsert(ctx context.Context, filter Filter, data T) (T, error)
	SilentUpdateMany(ctx context.Context, filter Filter, updatedData map[string]any) error
	SilentUpdateById(ctx context.Context, id ID, updatedData T, opts ...UpdateOpts) (T, error)
	DeleteById(ctx context.Context, id ID) error
	DeleteMany(ctx context.Context, filter Filter) error
	IndexFields(ctx context.Context, indices []IndexField) error
	// Delete(ctx context.Context, query Query) ([]ID, error)
	DeleteOne(ctx context.Context, filter Filter) error

	ErrAlreadyExists(err error) bool
	MergeMatchFilters(filter Filter, matchFilters map[string]MatchFilter) Filter
}

type indexOrder bool

const (
	IndexAsc  indexOrder = true
	IndexDesc indexOrder = false
)

type IndexKey struct {
	Key   string
	Value indexOrder
}

type IndexField struct {
	Field  []IndexKey
	Unique bool
}

func CursorFromBase64(b string) (Cursor, error) {
	b2, err := base64.StdEncoding.DecodeString(b)
	if err != nil {
		return Cursor(""), err
	}
	return Cursor(b2), nil
}

func CursorToBase64(c Cursor) string {
	return base64.StdEncoding.EncodeToString([]byte(c))
}

type CursorSortBy struct {
	Field     string        `json:"field"`
	Direction SortDirection `json:"sortDirection"`
}

type Cursor string

type CursorPagination struct {
	First *int64  `json:"first"`
	After *string `json:"after,omitempty"`

	Last   *int64  `json:"last,omitempty"`
	Before *string `json:"before,omitempty"`

	OrderBy       string        `json:"orderBy,omitempty" graphql:"default=\"_id\""`
	SortDirection SortDirection `json:"sortDirection,omitempty" graphql:"enum=ASC;DESC,default=\"ASC\""`
}

type SortDirection string

const (
	SortDirectionAsc  SortDirection = "ASC"
	SortDirectionDesc SortDirection = "DESC"
)

var DefaultCursorPagination = CursorPagination{
	First:         functions.New(int64(10)),
	After:         nil,
	OrderBy:       "_id",
	SortDirection: SortDirectionAsc,
}
