package repos

import (
	"context"
	t "kloudlite.io/pkg/types"
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
}

type Opts map[string]interface{}
type SortOpts map[string]int32
type Filter map[string]interface{}
type Query struct {
	Filter Filter
	Sort   map[string]interface{}
}

type ID string

type PageInfo struct {
	StartCursor string
	EndCursor   string
	HasNextPage bool
	HasPrevPage bool
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
	FindPaginated(ctx context.Context, filter Filter, pagination t.CursorPagination) (*PaginatedRecord[T], error)
	FindById(ctx context.Context, id ID) (T, error)
	Create(ctx context.Context, data T) (T, error)
	Exists(ctx context.Context, filter Filter) (bool, error)

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
