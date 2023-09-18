package mocks

import (
	context "context"
	repos "kloudlite.io/pkg/repos"
)

type DbRepoCallerInfo struct {
	Args []any
}

type DbRepo[T repos.Entity] struct {
	Calls                 map[string][]DbRepoCallerInfo
	MockCreate            func(ctx context.Context, data T) (T, error)
	MockDeleteById        func(ctx context.Context, id repos.ID) error
	MockDeleteMany        func(ctx context.Context, filter repos.Filter) error
	MockDeleteOne         func(ctx context.Context, filter repos.Filter) error
	MockErrAlreadyExists  func(err error) bool
	MockExists            func(ctx context.Context, filter repos.Filter) (bool, error)
	MockFind              func(ctx context.Context, query repos.Query) ([]T, error)
	MockFindById          func(ctx context.Context, id repos.ID) (T, error)
	MockFindOne           func(ctx context.Context, filter repos.Filter) (T, error)
	MockFindPaginated     func(ctx context.Context, filter repos.Filter, pagination repos.CursorPagination) (*repos.PaginatedRecord[T], error)
	MockIndexFields       func(ctx context.Context, indices []repos.IndexField) error
	MockMergeMatchFilters func(filter repos.Filter, matchFilters map[string]repos.MatchFilter) repos.Filter
	MockNewId             func() repos.ID
	MockSilentUpdateById  func(ctx context.Context, id repos.ID, updatedData T, opts ...repos.UpdateOpts) (T, error)
	MockSilentUpdateMany  func(ctx context.Context, filter repos.Filter, updatedData map[string]any) error
	MockSilentUpsert      func(ctx context.Context, filter repos.Filter, data T) (T, error)
	MockUpdateById        func(ctx context.Context, id repos.ID, updatedData T, opts ...repos.UpdateOpts) (T, error)
	MockUpdateMany        func(ctx context.Context, filter repos.Filter, updatedData map[string]any) error
	MockUpdateOne         func(ctx context.Context, filter repos.Filter, updatedData T, opts ...repos.UpdateOpts) (T, error)
	MockUpsert            func(ctx context.Context, filter repos.Filter, data T) (T, error)
}

func (m *DbRepo[T]) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]DbRepoCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], DbRepoCallerInfo{Args: args})
}

func (dMock *DbRepo[T]) Create(ctx context.Context, data T) (T, error) {
	if dMock.MockCreate != nil {
		dMock.registerCall("Create", ctx, data)
		return dMock.MockCreate(ctx, data)
	}
	panic("method 'Create' not implemented, yet")
}

func (dMock *DbRepo[T]) DeleteById(ctx context.Context, id repos.ID) error {
	if dMock.MockDeleteById != nil {
		dMock.registerCall("DeleteById", ctx, id)
		return dMock.MockDeleteById(ctx, id)
	}
	panic("method 'DeleteById' not implemented, yet")
}

func (dMock *DbRepo[T]) DeleteMany(ctx context.Context, filter repos.Filter) error {
	if dMock.MockDeleteMany != nil {
		dMock.registerCall("DeleteMany", ctx, filter)
		return dMock.MockDeleteMany(ctx, filter)
	}
	panic("method 'DeleteMany' not implemented, yet")
}

func (dMock *DbRepo[T]) DeleteOne(ctx context.Context, filter repos.Filter) error {
	if dMock.MockDeleteOne != nil {
		dMock.registerCall("DeleteOne", ctx, filter)
		return dMock.MockDeleteOne(ctx, filter)
	}
	panic("method 'DeleteOne' not implemented, yet")
}

func (dMock *DbRepo[T]) ErrAlreadyExists(err error) bool {
	if dMock.MockErrAlreadyExists != nil {
		dMock.registerCall("ErrAlreadyExists", err)
		return dMock.MockErrAlreadyExists(err)
	}
	panic("method 'ErrAlreadyExists' not implemented, yet")
}

func (dMock *DbRepo[T]) Exists(ctx context.Context, filter repos.Filter) (bool, error) {
	if dMock.MockExists != nil {
		dMock.registerCall("Exists", ctx, filter)
		return dMock.MockExists(ctx, filter)
	}
	panic("method 'Exists' not implemented, yet")
}

func (dMock *DbRepo[T]) Find(ctx context.Context, query repos.Query) ([]T, error) {
	if dMock.MockFind != nil {
		dMock.registerCall("Find", ctx, query)
		return dMock.MockFind(ctx, query)
	}
	panic("method 'Find' not implemented, yet")
}

func (dMock *DbRepo[T]) FindById(ctx context.Context, id repos.ID) (T, error) {
	if dMock.MockFindById != nil {
		dMock.registerCall("FindById", ctx, id)
		return dMock.MockFindById(ctx, id)
	}
	panic("method 'FindById' not implemented, yet")
}

func (dMock *DbRepo[T]) FindOne(ctx context.Context, filter repos.Filter) (T, error) {
	if dMock.MockFindOne != nil {
		dMock.registerCall("FindOne", ctx, filter)
		return dMock.MockFindOne(ctx, filter)
	}
	panic("method 'FindOne' not implemented, yet")
}

func (dMock *DbRepo[T]) FindPaginated(ctx context.Context, filter repos.Filter, pagination repos.CursorPagination) (*repos.PaginatedRecord[T], error) {
	if dMock.MockFindPaginated != nil {
		dMock.registerCall("FindPaginated", ctx, filter, pagination)
		return dMock.MockFindPaginated(ctx, filter, pagination)
	}
	panic("method 'FindPaginated' not implemented, yet")
}

func (dMock *DbRepo[T]) IndexFields(ctx context.Context, indices []repos.IndexField) error {
	if dMock.MockIndexFields != nil {
		dMock.registerCall("IndexFields", ctx, indices)
		return dMock.MockIndexFields(ctx, indices)
	}
	panic("method 'IndexFields' not implemented, yet")
}

func (dMock *DbRepo[T]) MergeMatchFilters(filter repos.Filter, matchFilters map[string]repos.MatchFilter) repos.Filter {
	if dMock.MockMergeMatchFilters != nil {
		dMock.registerCall("MergeMatchFilters", filter, matchFilters)
		return dMock.MockMergeMatchFilters(filter, matchFilters)
	}
	panic("method 'MergeMatchFilters' not implemented, yet")
}

func (dMock *DbRepo[T]) NewId() repos.ID {
	if dMock.MockNewId != nil {
		dMock.registerCall("NewId")
		return dMock.MockNewId()
	}
	panic("method 'NewId' not implemented, yet")
}

func (dMock *DbRepo[T]) SilentUpdateById(ctx context.Context, id repos.ID, updatedData T, opts ...repos.UpdateOpts) (T, error) {
	if dMock.MockSilentUpdateById != nil {
		dMock.registerCall("SilentUpdateById", ctx, id, updatedData, opts)
		return dMock.MockSilentUpdateById(ctx, id, updatedData, opts...)
	}
	panic("method 'SilentUpdateById' not implemented, yet")
}

func (dMock *DbRepo[T]) SilentUpdateMany(ctx context.Context, filter repos.Filter, updatedData map[string]any) error {
	if dMock.MockSilentUpdateMany != nil {
		dMock.registerCall("SilentUpdateMany", ctx, filter, updatedData)
		return dMock.MockSilentUpdateMany(ctx, filter, updatedData)
	}
	panic("method 'SilentUpdateMany' not implemented, yet")
}

func (dMock *DbRepo[T]) SilentUpsert(ctx context.Context, filter repos.Filter, data T) (T, error) {
	if dMock.MockSilentUpsert != nil {
		dMock.registerCall("SilentUpsert", ctx, filter, data)
		return dMock.MockSilentUpsert(ctx, filter, data)
	}
	panic("method 'SilentUpsert' not implemented, yet")
}

func (dMock *DbRepo[T]) UpdateById(ctx context.Context, id repos.ID, updatedData T, opts ...repos.UpdateOpts) (T, error) {
	if dMock.MockUpdateById != nil {
		dMock.registerCall("UpdateById", ctx, id, updatedData, opts)
		return dMock.MockUpdateById(ctx, id, updatedData, opts...)
	}
	panic("method 'UpdateById' not implemented, yet")
}

func (dMock *DbRepo[T]) UpdateMany(ctx context.Context, filter repos.Filter, updatedData map[string]any) error {
	if dMock.MockUpdateMany != nil {
		dMock.registerCall("UpdateMany", ctx, filter, updatedData)
		return dMock.MockUpdateMany(ctx, filter, updatedData)
	}
	panic("method 'UpdateMany' not implemented, yet")
}

func (dMock *DbRepo[T]) UpdateOne(ctx context.Context, filter repos.Filter, updatedData T, opts ...repos.UpdateOpts) (T, error) {
	if dMock.MockUpdateOne != nil {
		dMock.registerCall("UpdateOne", ctx, filter, updatedData, opts)
		return dMock.MockUpdateOne(ctx, filter, updatedData, opts...)
	}
	panic("method 'UpdateOne' not implemented, yet")
}

func (dMock *DbRepo[T]) Upsert(ctx context.Context, filter repos.Filter, data T) (T, error) {
	if dMock.MockUpsert != nil {
		dMock.registerCall("Upsert", ctx, filter, data)
		return dMock.MockUpsert(ctx, filter, data)
	}
	panic("method 'Upsert' not implemented, yet")
}

func NewDbRepo[T repos.Entity]() *DbRepo[T] {
	return &DbRepo[T]{}
}
