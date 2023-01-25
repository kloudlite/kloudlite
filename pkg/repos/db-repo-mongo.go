package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/fx"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
)

type dbRepo[T Entity] struct {
	db             *mongo.Database
	collectionName string
	shortName      string
}

var re = regexp.MustCompile(`(\W|_)+/g`)

func (repo *dbRepo[T]) NewId() ID {
	id, e := fn.CleanerNanoid(28)
	if e != nil {
		panic(fmt.Errorf("could not get cleanerNanoid()"))
	}
	return ID(fmt.Sprintf("%s-%s", repo.shortName, strings.ToLower(id)))
}
func (repo *dbRepo[T]) Find(ctx context.Context, query Query) ([]T, error) {
	results := make([]T, 0)
	curr, err := repo.db.Collection(repo.collectionName).Find(
		ctx, query.Filter, &options.FindOptions{
			Sort: query.Sort,
		},
	)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return results, nil
		}
		return nil, err
	}
	err = curr.All(ctx, &results)
	if err != nil {
		return nil, err
	}
	return results, err
}

func (repo *dbRepo[T]) findOne(ctx context.Context, filter Filter) (T, error) {
	one := repo.db.Collection(repo.collectionName).FindOne(ctx, filter)
	res := fn.NewTypeFromPointer[T]()
	err := one.Decode(&res)
	if err != nil {
		fmt.Println("(debug/info) ERR: ", err)
		return res, err
	}
	return res, nil
}
func (repo *dbRepo[T]) FindOne(ctx context.Context, filter Filter) (T, error) {
	one := repo.db.Collection(repo.collectionName).FindOne(ctx, filter)
	t := make([]T, 1)
	// res := fn.NewTypeFromPointer[T]()
	err := one.Decode(&t[0])
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return t[0], nil
		}
		return t[0], err
	}
	return t[0], nil
}
func (repo *dbRepo[T]) FindPaginated(ctx context.Context, query Query, page int64, size int64, opts ...Opts) (PaginatedRecord[T], error) {
	results := make([]T, 0)
	var offset int64 = (page - 1) * size
	curr, e := repo.db.Collection(repo.collectionName).Find(
		ctx, query.Filter, &options.FindOptions{
			Limit: &size,
			Skip:  &offset,
			Sort:  query.Sort,
		},
	)
	e = curr.All(ctx, results)

	total, e := repo.db.Collection(repo.collectionName).CountDocuments(ctx, query.Filter)

	return PaginatedRecord[T]{
		Results:    results,
		TotalCount: total,
	}, e
}
func (repo *dbRepo[T]) FindById(ctx context.Context, id ID) (T, error) {
	var result T
	r := repo.db.Collection(repo.collectionName).FindOne(ctx, &Filter{"id": id})

	err := r.Decode(&result)
	if err != nil {
		return result, err
	}
	return result, err
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
	repo.withCreationTime(data)

	b, err := json.Marshal(data)
	if err != nil {
		var x T
		return x, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		var x T
		return x, err
	}

	r, e := repo.db.Collection(repo.collectionName).InsertOne(ctx, m)
	if e != nil {
		var x T
		return x, e
	}
	var result T
	r2 := repo.db.Collection(repo.collectionName).FindOne(ctx, Filter{"_id": r.InsertedID})

	var resultM map[string]any
	e = r2.Decode(&resultM)

	m2, err := json.Marshal(resultM)
	if err := json.Unmarshal(m2, &result); err != nil {
		var x T
		return x, err
	}

	return result, e
}

func (repo *dbRepo[T]) Exists(ctx context.Context, filter Filter) (bool, error) {
	one := repo.db.Collection(repo.collectionName).FindOne(ctx, filter)
	res := fn.NewTypeFromPointer[T]()
	err := one.Decode(&res)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
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
		return err
	}
	return nil
}

func (repo *dbRepo[T]) UpdateOne(ctx context.Context, filter Filter, updatedData T, opts ...UpdateOpts) (T, error) {
	var result T
	after := options.After
	updateOpts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	if opt := fn.ParseOnlyOption[UpdateOpts](opts); opt != nil {
		updateOpts.Upsert = &opt.Upsert
	}
	repo.withUpdateTime(updatedData)
	r := repo.db.Collection(repo.collectionName).FindOneAndUpdate(
		ctx,
		filter,
		bson.M{"$set": updatedData},
		updateOpts,
	)
	e := r.Decode(&result)
	return result, e
}

func (repo *dbRepo[T]) UpdateById(ctx context.Context, id ID, updatedData T, opts ...UpdateOpts) (T, error) {
	var result T
	after := options.After
	updateOpts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	if opt := fn.ParseOnlyOption[UpdateOpts](opts); opt != nil {
		updateOpts.Upsert = &opt.Upsert
	}
	repo.withUpdateTime(updatedData)
	r := repo.db.Collection(repo.collectionName).FindOneAndUpdate(
		ctx,
		&Filter{"id": id},
		bson.M{"$set": updatedData},
		updateOpts,
	)
	e := r.Decode(&result)
	return result, e
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
	e := r.Decode(&result)
	return e
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
	return err
}

func (repo *dbRepo[T]) DeleteMany(ctx context.Context, filter Filter) error {
	_, err := repo.db.Collection(repo.collectionName).DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

func (repo *dbRepo[T]) IndexFields(ctx context.Context, indices []IndexField) error {
	if len(indices) == 0 {
		return nil
	}
	var models []mongo.IndexModel
	for _, f := range indices {
		b := bson.D{}
		for _, field := range f.Field {
			switch field.Value {
			case IndexAsc:
				b = append(b, bson.E{Key: field.Key, Value: 1})
			case IndexDesc:
				b = append(b, bson.E{Key: field.Key, Value: -1})
			}
		}
		models = append(
			models, mongo.IndexModel{
				Keys: b,
				Options: &options.IndexOptions{
					Unique: &f.Unique,
				},
			},
		)
	}
	_, err := repo.db.Collection(repo.collectionName).Indexes().CreateMany(ctx, models)
	return err
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
		return err
	}
	return nil
}
func (repo *dbRepo[T]) SilentUpdateById(ctx context.Context, id ID, updatedData T, opts ...UpdateOpts) (T, error) {
	var result T
	after := options.After
	updateOpts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}
	if opt := fn.ParseOnlyOption[UpdateOpts](opts); opt != nil {
		updateOpts.Upsert = &opt.Upsert
	}
	r := repo.db.Collection(repo.collectionName).FindOneAndUpdate(
		ctx,
		&Filter{"id": id},
		bson.M{"$set": updatedData},
		updateOpts,
	)
	e := r.Decode(&result)
	return result, e
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
			func(lifecycle fx.Lifecycle, repo DbRepo[T]) {
				lifecycle.Append(
					fx.Hook{
						OnStart: func(ctx context.Context) error {
							err := repo.IndexFields(ctx, indexFields)
							if err != nil {
								return errors.NewEf(err, "could not create indexes on DB for repo %T", repo)
							}
							return nil
						},
					},
				)
			},
		),
	)
}
