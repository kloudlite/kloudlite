package mocks

import (
	context "context"
	repos "github.com/kloudlite/api/pkg/repos"
)

type DbRepoCallerInfo struct {
	Args []any
}

type DbRepo[T repos.Entity] struct {
	Calls                      map[string][]DbRepoCallerInfo
	MockCount                  func(ctx context.Context, filter repos.Filter) (int64, error)
	MockCreate                 func(ctx context.Context, data T) (T, error)
	MockCreateMany             func(ctx context.Context, data []T) error
	MockDeleteById             func(ctx context.Context, id repos.ID) error
	MockDeleteMany             func(ctx context.Context, filter repos.Filter) error
	MockDeleteOne              func(ctx context.Context, filter repos.Filter) error
	MockErrAlreadyExists       func(err error) bool
	MockExists                 func(ctx context.Context, filter repos.Filter) (bool, error)
	MockFind                   func(ctx context.Context, query repos.Query) ([]T, error)
	MockFindById               func(ctx context.Context, id repos.ID) (T, error)
	MockFindOne                func(ctx context.Context, filter repos.Filter) (T, error)
	MockFindPaginated          func(ctx context.Context, filter repos.Filter, pagination repos.CursorPagination) (*repos.PaginatedRecord[T], error)
	MockIndexFields            func(ctx context.Context, indices []repos.IndexField) error
	MockMergeMatchFilters      func(filter repos.Filter, matchFilters map[string]repos.MatchFilter) repos.Filter
	MockNewId                  func() repos.ID
	MockPatch                  func(ctx context.Context, filter repos.Filter, patch repos.Document, opts ...repos.UpdateOpts) (T, error)
	MockPatchById              func(ctx context.Context, id repos.ID, patch repos.Document, opts ...repos.UpdateOpts) (T, error)
	MockPatchOne               func(ctx context.Context, filter repos.Filter, patch repos.Document, opts ...repos.UpdateOpts) (T, error)
	MockUpdateById             func(ctx context.Context, id repos.ID, updatedData T, opts ...repos.UpdateOpts) (T, error)
	MockUpdateMany             func(ctx context.Context, filter repos.Filter, updatedData map[string]any) error
	MockUpdateOne              func(ctx context.Context, filter repos.Filter, updatedData T, opts ...repos.UpdateOpts) (T, error)
	MockUpdateWithVersionCheck func(ctx context.Context, id repos.ID, updatedData T) (T, error)
	MockUpsert                 func(ctx context.Context, filter repos.Filter, data T) (T, error)
}

func (m *DbRepo[T]) registerCall(funcName string, args ...any) {
	if m.Calls == nil {
		m.Calls = map[string][]DbRepoCallerInfo{}
	}
	m.Calls[funcName] = append(m.Calls[funcName], DbRepoCallerInfo{Args: args})
}

func (dMock *DbRepo[T]) Count(ctx context.Context, filter repos.Filter) (int64, error) {
	if dMock.MockCount != nil {
		dMock.registerCall("Count", ctx, filter)
		return dMock.MockCount(ctx, filter)
	}
	panic("DbRepo[T]: method 'Count' not implemented, yet")
}

func (dMock *DbRepo[T]) Create(ctx context.Context, data T) (T, error) {
	if dMock.MockCreate != nil {
		dMock.registerCall("Create", ctx, data)
		return dMock.MockCreate(ctx, data)
	}
	panic("DbRepo[T]: method 'Create' not implemented, yet")
}

func (dMock *DbRepo[T]) CreateMany(ctx context.Context, data []T) error {
	if dMock.MockCreateMany != nil {
		dMock.registerCall("CreateMany", ctx, data)
		return dMock.MockCreateMany(ctx, data)
	}
	panic("DbRepo[T]: method 'CreateMany' not implemented, yet")
}

func (dMock *DbRepo[T]) DeleteById(ctx context.Context, id repos.ID) error {
	if dMock.MockDeleteById != nil {
		dMock.registerCall("DeleteById", ctx, id)
		return dMock.MockDeleteById(ctx, id)
	}
	panic("DbRepo[T]: method 'DeleteById' not implemented, yet")
}

func (dMock *DbRepo[T]) DeleteMany(ctx context.Context, filter repos.Filter) error {
	if dMock.MockDeleteMany != nil {
		dMock.registerCall("DeleteMany", ctx, filter)
		return dMock.MockDeleteMany(ctx, filter)
	}
	panic("DbRepo[T]: method 'DeleteMany' not implemented, yet")
}

func (dMock *DbRepo[T]) DeleteOne(ctx context.Context, filter repos.Filter) error {
	if dMock.MockDeleteOne != nil {
		dMock.registerCall("DeleteOne", ctx, filter)
		return dMock.MockDeleteOne(ctx, filter)
	}
	panic("DbRepo[T]: method 'DeleteOne' not implemented, yet")
}

func (dMock *DbRepo[T]) ErrAlreadyExists(err error) bool {
	if dMock.MockErrAlreadyExists != nil {
		dMock.registerCall("ErrAlreadyExists", err)
		return dMock.MockErrAlreadyExists(err)
	}
	panic("DbRepo[T]: method 'ErrAlreadyExists' not implemented, yet")
}

func (dMock *DbRepo[T]) Exists(ctx context.Context, filter repos.Filter) (bool, error) {
	if dMock.MockExists != nil {
		dMock.registerCall("Exists", ctx, filter)
		return dMock.MockExists(ctx, filter)
	}
	panic("DbRepo[T]: method 'Exists' not implemented, yet")
}

func (dMock *DbRepo[T]) Find(ctx context.Context, query repos.Query) ([]T, error) {
	if dMock.MockFind != nil {
		dMock.registerCall("Find", ctx, query)
		return dMock.MockFind(ctx, query)
	}
	panic("DbRepo[T]: method 'Find' not implemented, yet")
}

func (dMock *DbRepo[T]) FindById(ctx context.Context, id repos.ID) (T, error) {
	if dMock.MockFindById != nil {
		dMock.registerCall("FindById", ctx, id)
		return dMock.MockFindById(ctx, id)
	}
	panic("DbRepo[T]: method 'FindById' not implemented, yet")
}

func (dMock *DbRepo[T]) FindOne(ctx context.Context, filter repos.Filter) (T, error) {
	if dMock.MockFindOne != nil {
		dMock.registerCall("FindOne", ctx, filter)
		return dMock.MockFindOne(ctx, filter)
	}
	panic("DbRepo[T]: method 'FindOne' not implemented, yet")
}

func (dMock *DbRepo[T]) FindPaginated(ctx context.Context, filter repos.Filter, pagination repos.CursorPagination) (*repos.PaginatedRecord[T], error) {
	if dMock.MockFindPaginated != nil {
		dMock.registerCall("FindPaginated", ctx, filter, pagination)
		return dMock.MockFindPaginated(ctx, filter, pagination)
	}
	panic("DbRepo[T]: method 'FindPaginated' not implemented, yet")
}

func (dMock *DbRepo[T]) IndexFields(ctx context.Context, indices []repos.IndexField) error {
	if dMock.MockIndexFields != nil {
		dMock.registerCall("IndexFields", ctx, indices)
		return dMock.MockIndexFields(ctx, indices)
	}
	panic("DbRepo[T]: method 'IndexFields' not implemented, yet")
}

func (dMock *DbRepo[T]) MergeMatchFilters(filter repos.Filter, matchFilters map[string]repos.MatchFilter) repos.Filter {
	if dMock.MockMergeMatchFilters != nil {
		dMock.registerCall("MergeMatchFilters", filter, matchFilters)
		return dMock.MockMergeMatchFilters(filter, matchFilters)
	}
	panic("DbRepo[T]: method 'MergeMatchFilters' not implemented, yet")
}

func (dMock *DbRepo[T]) NewId() repos.ID {
	if dMock.MockNewId != nil {
		dMock.registerCall("NewId")
		return dMock.MockNewId()
	}
	panic("DbRepo[T]: method 'NewId' not implemented, yet")
}

func (dMock *DbRepo[T]) Patch(ctx context.Context, filter repos.Filter, patch repos.Document, opts ...repos.UpdateOpts) (T, error) {
	if dMock.MockPatch != nil {
		dMock.registerCall("Patch", ctx, filter, patch, opts)
		return dMock.MockPatch(ctx, filter, patch, opts...)
	}
	panic("DbRepo[T]: method 'Patch' not implemented, yet")
}

func (dMock *DbRepo[T]) PatchById(ctx context.Context, id repos.ID, patch repos.Document, opts ...repos.UpdateOpts) (T, error) {
	if dMock.MockPatchById != nil {
		dMock.registerCall("PatchById", ctx, id, patch, opts)
		return dMock.MockPatchById(ctx, id, patch, opts...)
	}
	panic("DbRepo[T]: method 'PatchById' not implemented, yet")
}

func (dMock *DbRepo[T]) PatchOne(ctx context.Context, filter repos.Filter, patch repos.Document, opts ...repos.UpdateOpts) (T, error) {
	if dMock.MockPatchOne != nil {
		dMock.registerCall("PatchOne", ctx, filter, patch, opts)
		return dMock.MockPatchOne(ctx, filter, patch, opts...)
	}
	panic("DbRepo[T]: method 'PatchOne' not implemented, yet")
}

func (dMock *DbRepo[T]) UpdateById(ctx context.Context, id repos.ID, updatedData T, opts ...repos.UpdateOpts) (T, error) {
	if dMock.MockUpdateById != nil {
		dMock.registerCall("UpdateById", ctx, id, updatedData, opts)
		return dMock.MockUpdateById(ctx, id, updatedData, opts...)
	}
	panic("DbRepo[T]: method 'UpdateById' not implemented, yet")
}

func (dMock *DbRepo[T]) UpdateMany(ctx context.Context, filter repos.Filter, updatedData map[string]any) error {
	if dMock.MockUpdateMany != nil {
		dMock.registerCall("UpdateMany", ctx, filter, updatedData)
		return dMock.MockUpdateMany(ctx, filter, updatedData)
	}
	panic("DbRepo[T]: method 'UpdateMany' not implemented, yet")
}

func (dMock *DbRepo[T]) UpdateOne(ctx context.Context, filter repos.Filter, updatedData T, opts ...repos.UpdateOpts) (T, error) {
	if dMock.MockUpdateOne != nil {
		dMock.registerCall("UpdateOne", ctx, filter, updatedData, opts)
		return dMock.MockUpdateOne(ctx, filter, updatedData, opts...)
	}
	panic("DbRepo[T]: method 'UpdateOne' not implemented, yet")
}

func (dMock *DbRepo[T]) UpdateWithVersionCheck(ctx context.Context, id repos.ID, updatedData T) (T, error) {
	if dMock.MockUpdateWithVersionCheck != nil {
		dMock.registerCall("UpdateWithVersionCheck", ctx, id, updatedData)
		return dMock.MockUpdateWithVersionCheck(ctx, id, updatedData)
	}
	panic("DbRepo[T]: method 'UpdateWithVersionCheck' not implemented, yet")
}

func (dMock *DbRepo[T]) Upsert(ctx context.Context, filter repos.Filter, data T) (T, error) {
	if dMock.MockUpsert != nil {
		dMock.registerCall("Upsert", ctx, filter, data)
		return dMock.MockUpsert(ctx, filter, data)
	}
	panic("DbRepo[T]: method 'Upsert' not implemented, yet")
}

func NewDbRepo[T repos.Entity]() *DbRepo[T] {
	return &DbRepo[T]{}
}
