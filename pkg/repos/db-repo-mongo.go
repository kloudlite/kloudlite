package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/fx"

	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type dbRepo[T Entity] struct {
	db             *mongo.Database
	collectionName string
	shortName      string
}

var re = regexp.MustCompile(`(\W|_)+/g`)

func toMap(v any) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, errors.NewE(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.NewE(err)
	}
	return m, nil
}

func fromMap[T Entity](v map[string]any) (T, error) {
	var emptyResult T
	b, err := json.Marshal(v)
	if err != nil {
		return emptyResult, errors.NewE(err)
	}
	var result T
	if err := json.Unmarshal(b, &result); err != nil {
		return emptyResult, errors.NewE(err)
	}
	return result, nil
}

func bsonToStruct[T any](r *mongo.SingleResult) (T, error) {
	var m map[string]any
	var result T
	if err := r.Decode(&m); err != nil {
		return result, errors.NewE(err)
	}
	b, err := json.Marshal(m)
	if err != nil {
		return result, errors.NewE(err)
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return result, errors.NewE(err)
	}
	return result, nil
}

func cursorToStruct[T any](ctx context.Context, curr *mongo.Cursor) ([]T, error) {
	var m []map[string]any
	var results []T

	if err := curr.All(ctx, &m); err != nil {
		return results, errors.NewE(err)
	}

	b, err := json.Marshal(m)
	if err != nil {
		return results, errors.NewE(err)
	}

	if err := json.Unmarshal(b, &results); err != nil {
		return results, errors.NewE(err)
	}

	return results, nil
}

func (repo *dbRepo[T]) NewId() ID {
	id, e := fn.CleanerNanoid(28)
	if e != nil {
		panic(errors.Newf("could not get cleanerNanoid()"))
	}
	return ID(fmt.Sprintf("%s-%s", repo.shortName, strings.ToLower(id)))
}

func (repo *dbRepo[T]) Find(ctx context.Context, query Query) ([]T, error) {
	curr, err := repo.db.Collection(repo.collectionName).Find(
		ctx, query.Filter, &options.FindOptions{
			Sort: query.Sort,
		},
	)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return make([]T, 0), nil
		}
		return nil, errors.NewE(err)
	}

	return cursorToStruct[T](ctx, curr)
}

func (repo *dbRepo[T]) Count(ctx context.Context, filter Filter) (int64, error) {
	return repo.db.Collection(repo.collectionName).CountDocuments(ctx, filter, options.Count().SetMaxTime(1*time.Second))
}

func (repo *dbRepo[T]) findOne(ctx context.Context, filter Filter) (T, error) {
	one := repo.db.Collection(repo.collectionName).FindOne(ctx, filter)
	item, err := bsonToStruct[T](one)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return item, errors.Newf("no document found")
		}
		return item, errors.NewE(err)
	}
	return item, nil
}

func (repo *dbRepo[T]) FindOne(ctx context.Context, filter Filter) (T, error) {
	one := repo.db.Collection(repo.collectionName).FindOne(ctx, filter)
	item, err := bsonToStruct[T](one)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return item, nil
		}
		return item, errors.NewE(err)
	}
	return item, nil
}

func (repo *dbRepo[T]) FindPaginated(ctx context.Context, filter Filter, pagination CursorPagination) (*PaginatedRecord[T], error) {
	if pagination.First != nil && pagination.Last != nil {
		return nil, errors.Newf("first/last only one of these parameters could be passed on, you have specified both")
	}

	if pagination.After != nil && pagination.Before != nil {
		return nil, errors.Newf("after/before only one of these parameters could be passed on, you have specified both")
	}

	if pagination.After != nil && pagination.First == nil {
		return nil, errors.Newf("paramter `after` requires paramter `first` to be specified")
	}

	if pagination.Before != nil && pagination.Last == nil {
		return nil, errors.Newf("paramter `before` requires paramter `last` to be specified")
	}

	queryFilter := Filter{}

	for k, v := range filter {
		queryFilter[k] = v
	}

	if pagination.After != nil {
		aft, err := CursorFromBase64(*pagination.After)
		if err != nil {
			return nil, errors.NewE(err)
		}
		objectID, err := primitive.ObjectIDFromHex(string(aft))
		if err != nil {
			return nil, errors.NewE(err)
		}
		queryFilter["_id"] = bson.M{"$gt": objectID}
	}

	if pagination.Before != nil {
		bef, err := CursorFromBase64(*pagination.Before)
		if err != nil {
			return nil, errors.NewE(err)
		}
		objectID, err := primitive.ObjectIDFromHex(string(bef))
		if err != nil {
			return nil, errors.NewE(err)
		}
		queryFilter["_id"] = bson.M{"$lt": objectID}
	}

	var limit int64
	if pagination.Last != nil {
		limit = *pagination.Last + 1
	}

	if pagination.First != nil {
		limit = *pagination.First + 1
	}

	// var results []T
	curr, err := repo.db.Collection(repo.collectionName).Find(
		ctx, queryFilter, &options.FindOptions{
			Limit: &limit,
			Sort: bson.M{pagination.OrderBy: func() int {
				if pagination.SortDirection == SortDirectionDesc {
					return -1
				}
				return 1
			}()},
		},
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	results, err := cursorToStruct[T](ctx, curr)
	if err != nil {
		return nil, errors.NewE(err)
	}

	total, err := repo.db.Collection(repo.collectionName).CountDocuments(ctx, filter)
	if err != nil {
		return nil, errors.NewE(err)
	}

	pageInfo := PageInfo{}

	if len(results) > 0 {
		pageInfo.StartCursor = CursorToBase64(Cursor(string(results[0].GetPrimitiveID())))
		pageInfo.EndCursor = CursorToBase64(Cursor(string(results[len(results)-1].GetPrimitiveID())))

		if pagination.First != nil {
			pageInfo.HasNextPage = fn.New(len(results) > int(*pagination.First))
			if pageInfo.HasNextPage != nil && *pageInfo.HasNextPage {
				results = results[:*pagination.First]
			}
			pageInfo.HasPrevPage = fn.New(pagination.After != nil)
		}

		if pagination.Last != nil {
			pageInfo.HasNextPage = fn.New(pagination.Before != nil)
			pageInfo.HasPrevPage = fn.New(len(results) > int(*pagination.Last))

			if pageInfo.HasPrevPage != nil && *pageInfo.HasPrevPage {
				results = results[:*pagination.Last]
			}
		}
	}

	edges := make([]RecordEdge[T], len(results))
	for i := range results {
		edges[i] = RecordEdge[T]{
			Node:   results[i],
			Cursor: CursorToBase64(Cursor(results[i].GetPrimitiveID())),
		}
	}

	return &PaginatedRecord[T]{
		Edges:      edges,
		PageInfo:   pageInfo,
		TotalCount: total,
	}, nil
}

func (repo *dbRepo[T]) FindById(ctx context.Context, id ID) (T, error) {
	r := repo.db.Collection(repo.collectionName).FindOne(ctx, &Filter{"id": id})
	return bsonToStruct[T](r)
}

func (repo *dbRepo[T]) withId(data T) {
	if data.GetId() != "" {
		return
	}
	data.SetId(repo.NewId())
}

func (repo *dbRepo[T]) withCreationTime(data T) {
	if !data.GetCreationTime().IsZero() {
		return
	}
	data.SetCreationTime(time.Now())
	data.SetUpdateTime(time.Now())
}

func (repo *dbRepo[T]) withUpdateTime(data T) {
	data.SetUpdateTime(time.Now())
}

func (repo *dbRepo[T]) Create(ctx context.Context, data T) (T, error) {
	repo.withId(data)
	data.SetCreationTime(time.Now())
	data.SetUpdateTime(time.Now())

	var emptyResult T

	inputM, err := toMap(data)
	if err != nil {
		return emptyResult, errors.NewE(err)
	}

	r, err := repo.db.Collection(repo.collectionName).InsertOne(ctx, inputM)
	if err != nil {
		return emptyResult, errors.NewE(err)
	}

	r2 := repo.db.Collection(repo.collectionName).FindOne(ctx, Filter{"_id": r.InsertedID})
	return bsonToStruct[T](r2)
}

func (repo *dbRepo[T]) Exists(ctx context.Context, filter Filter) (bool, error) {
	one := repo.db.Collection(repo.collectionName).FindOne(ctx, filter)
	var m map[string]any
	err := one.Decode(&m)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, errors.NewE(err)
	}
	return true, nil
}

func (repo *dbRepo[T]) UpdateMany(ctx context.Context, filter Filter, updatedData map[string]any) error {
	updatedData["updateTime"] = time.Now()

	_, err := repo.db.Collection(repo.collectionName).UpdateMany(
		ctx,
		filter,
		bson.M{"$set": updatedData},
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (repo *dbRepo[T]) UpdateOne(ctx context.Context, filter Filter, updatedData T, opts ...UpdateOpts) (T, error) {
	after := options.After
	updateOpts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	if opt := fn.ParseOnlyOption[UpdateOpts](opts); opt != nil {
		updateOpts.Upsert = &opt.Upsert
	}

	updatedData.SetUpdateTime(time.Now())

	m, err := toMap(updatedData)
	if err != nil {
		var x T
		return x, errors.NewE(err)
	}

	r := repo.db.Collection(repo.collectionName).FindOneAndUpdate(
		ctx,
		filter,
		bson.M{"$set": m},
		updateOpts,
	)

	return bsonToStruct[T](r)
}

func (repo *dbRepo[T]) UpdateById(ctx context.Context, id ID, updatedData T, opts ...UpdateOpts) (T, error) {
	var result T
	after := options.After
	updateOpts := &options.FindOneAndUpdateOptions{ReturnDocument: &after}

	if opt := fn.ParseOnlyOption[UpdateOpts](opts); opt != nil {
		updateOpts.Upsert = &opt.Upsert
	}

	updatedData.SetUpdateTime(time.Now())

	m, err := toMap(updatedData)
	if err != nil {
		return result, errors.NewE(err)
	}

	delete(m, "_id")

	r := repo.db.Collection(repo.collectionName).FindOneAndUpdate(
		ctx,
		&Filter{"id": id},
		bson.M{"$set": m},
		updateOpts,
	)
	return bsonToStruct[T](r)
}

func (repo *dbRepo[T]) Upsert(ctx context.Context, filter Filter, data T) (T, error) {
	id := func() ID {
		if data.GetId() != "" {
			return data.GetId()
		}
		if t, err := repo.findOne(ctx, filter); err == nil {
			return t.GetId()
		}
		return repo.NewId()
	}()

	data.SetId(id)
	repo.withCreationTime(data)
	repo.withUpdateTime(data)

	return repo.UpdateById(
		ctx, id, data, UpdateOpts{
			Upsert: true,
		},
	)
}

func (repo *dbRepo[T]) DeleteById(ctx context.Context, id ID) error {
	var result T
	r := repo.db.Collection(repo.collectionName).FindOneAndDelete(ctx, &Filter{"id": id})
	return r.Decode(&result)
}

func (repo *dbRepo[T]) DeleteOne(ctx context.Context, filter Filter) error {
	var result T
	r := repo.db.Collection(repo.collectionName).FindOneAndDelete(ctx, filter)
	err := r.Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
	}
	return errors.NewE(err)
}

func (repo *dbRepo[T]) DeleteMany(ctx context.Context, filter Filter) error {
	_, err := repo.db.Collection(repo.collectionName).DeleteMany(ctx, filter)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}

func buildIndexName(curr string, indexKey string, indexValue int) string {
	if curr == "" {
		return curr + fmt.Sprintf("%s_%d", indexKey, indexValue)
	}
	return curr + "_" + fmt.Sprintf("%s_%d", indexKey, indexValue)
}

func (repo *dbRepo[T]) IndexFields(ctx context.Context, indices []IndexField) error {
	if len(indices) == 0 {
		return nil
	}

	indices = append(indices, IndexField{
		Field:  []IndexKey{{Key: "creationTime", Value: IndexAsc}},
		Unique: false,
	})
	// var models []mongo.IndexModel
	for _, f := range indices {
		b := bson.D{}

		// This method to create indexes ensure temporary safety, while index modification,
		// which we need to do when index names clash
		// READ MORE @ https://www.mongodb.com/docs/manual/tutorial/manage-indexes/#modify-an-index
		indexName := ""
		for _, field := range f.Field {
			switch field.Value {
			case IndexAsc:
				b = append(b, bson.E{Key: field.Key, Value: 1})
				indexName = buildIndexName(indexName, field.Key, 1)
			case IndexDesc:
				b = append(b, bson.E{Key: field.Key, Value: -1})
				indexName = buildIndexName(indexName, field.Key, -1)
			}
		}

		indexModel := mongo.IndexModel{Keys: b, Options: &options.IndexOptions{Unique: &f.Unique, Name: &indexName}}

		_, err := repo.db.Collection(repo.collectionName).Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			dummyKey := fn.CleanerNanoidOrDie(10)
			b2 := append(b, bson.E{Key: dummyKey, Value: 1})
			dummyIdxName := buildIndexName(indexName, dummyKey, 1)
			_, err := repo.db.Collection(repo.collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{Keys: b2, Options: &options.IndexOptions{Unique: &f.Unique, Name: &dummyIdxName}})
			if err != nil {
				return errors.NewE(err)
			}
			_, err = repo.db.Collection(repo.collectionName).Indexes().DropOne(ctx, indexName)
			if err != nil {
				return errors.NewE(err)
			}
			if _, err := repo.db.Collection(repo.collectionName).Indexes().CreateOne(ctx, indexModel); err != nil {
				return errors.NewE(err)
			}
			if _, err := repo.db.Collection(repo.collectionName).Indexes().DropOne(ctx, dummyIdxName); err != nil {
				return errors.NewE(err)
			}
		}
	}
	return nil
}

func (repo *dbRepo[T]) SilentUpsert(ctx context.Context, filter Filter, data T) (T, error) {
	id := func() ID {
		if data.GetId() != "" {
			return data.GetId()
		}
		if t, err := repo.findOne(ctx, filter); err == nil {
			return t.GetId()
		}
		return repo.NewId()
	}()
	data.SetId(id)
	if data.GetCreationTime().IsZero() {
		repo.withCreationTime(data)
	}
	return repo.UpdateById(
		ctx, id, data, UpdateOpts{
			Upsert: true,
		},
	)
}

func (repo *dbRepo[T]) SilentUpdateMany(ctx context.Context, filter Filter, updatedData map[string]any) error {
	_, err := repo.db.Collection(repo.collectionName).UpdateMany(
		ctx,
		filter,
		bson.M{"$set": updatedData},
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (repo *dbRepo[T]) SilentUpdateById(ctx context.Context, id ID, updatedData T, opts ...UpdateOpts) (T, error) {
	after := options.After
	updateOpts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	if opt := fn.ParseOnlyOption[UpdateOpts](opts); opt != nil {
		updateOpts.Upsert = &opt.Upsert
	}

	m, err := toMap(updatedData)
	if err != nil {
		var x T
		return x, errors.NewE(err)
	}

	r := repo.db.Collection(repo.collectionName).FindOneAndUpdate(
		ctx,
		&Filter{"id": id},
		bson.M{"$set": m},
		updateOpts,
	)
	return bsonToStruct[T](r)
}

func (repo *dbRepo[T]) ErrAlreadyExists(err error) bool {
	return mongo.IsDuplicateKeyError(err)
}

func (repo *dbRepo[T]) MergeMatchFilters(filter Filter, mFilter map[string]MatchFilter) Filter {
	if filter == nil {
		filter = map[string]any{}
	}

	for k, v := range mFilter {
		_, ok := filter[k]
		if ok {
			fmt.Printf("skipping search filter field %q, as it is already specified in filter", k)
			continue
		}
		switch v.MatchType {
		case MatchTypeExact:
			filter[k] = v.Exact
		case MatchTypeArray:
			filter[k] = bson.M{"$in": v.Array}
		case MatchTypeRegex:
			filter[k] = bson.M{"$regex": primitive.Regex{Pattern: *v.Regex, Options: "i"}}
		}
	}
	return filter
}

type MongoRepoOptions struct {
	IndexFields []string
}

func NewMongoRepo[T Entity](
	db *mongo.Database,
	collectionName string,
	shortName string,
) DbRepo[T] {
	return &dbRepo[T]{
		db,
		collectionName,
		shortName,
	}
}

func NewFxMongoRepo[T Entity](collectionName, shortName string, indexFields []IndexField) fx.Option {
	return fx.Module(
		"repo",
		fx.Provide(
			func(db *mongo.Database) DbRepo[T] {
				return NewMongoRepo[T](
					db,
					collectionName,
					shortName,
				)
			},
		),
		fx.Invoke(
			func(lifecycle fx.Lifecycle, repo DbRepo[T], logger logging.Logger) {
				lifecycle.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							go func() {
								err := repo.IndexFields(ctx, indexFields)
								if err != nil {
									logger.Errorf(err, "failed to update indexes on DB for repo %T", repo)
								}
								logger.Infof("indexes updated on DB for repo %T", repo)
							}()
							return nil
						},
					},
				)
			},
		),
	)
}
