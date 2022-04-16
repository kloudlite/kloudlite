package repos

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/fx"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kloudlite.io/pkg/functions"
)

type dbRepo[T Entity] struct {
	db             *mongo.Database
	collectionName string
	shortName      string
	options        *MongoRepoOptions
}

var re = regexp.MustCompile(`(\W|_)+/g`)

func (repo dbRepo[T]) NewId() ID {
	id, e := functions.CleanerNanoid(28)
	if e != nil {
		panic(fmt.Errorf("could not get cleanerNanoid()"))
	}
	return ID(fmt.Sprintf("%s-%s", repo.shortName, strings.ToLower(id)))
}

func (repo dbRepo[T]) Find(ctx context.Context, query Query) ([]T, error) {
	results := make([]T, 0)
	curr, err := repo.db.Collection(repo.collectionName).Find(ctx, query.Filter, &options.FindOptions{
		Sort: query.Sort,
	})
	err = curr.All(ctx, &results)
	return results, err
}

func (repo dbRepo[T]) FindPaginated(ctx context.Context, query Query, page int64, size int64, opts ...Opts) (PaginatedRecord[T], error) {
	results := make([]T, 0)
	var offset int64 = (page - 1) * size
	curr, e := repo.db.Collection(repo.collectionName).Find(ctx, query.Filter, &options.FindOptions{
		Limit: &size,
		Skip:  &offset,
		Sort:  query.Sort,
	})
	e = curr.All(ctx, results)

	total, e := repo.db.Collection(repo.collectionName).CountDocuments(ctx, query.Filter)

	return PaginatedRecord[T]{
		results:    results,
		totalCount: total,
	}, e
}

func (repo dbRepo[T]) FindById(ctx context.Context, id ID) (T, error) {
	var result T
	r := repo.db.Collection(repo.collectionName).FindOne(ctx, &Filter{"id": id})
	err := r.Decode(&result)
	return result, err
}

func (repo dbRepo[T]) withId(data T) T {
	if data.GetId() != "" {
		return data
	}
	data.SetId(repo.NewId())
	return data
}

func (repo dbRepo[T]) Create(ctx context.Context, data T) (T, error) {
	var result T
	recordWithId := repo.withId(data)
	r, e := repo.db.Collection(repo.collectionName).InsertOne(ctx, recordWithId)
	if e != nil {
		return result, e
	}
	r2 := repo.db.Collection(repo.collectionName).FindOne(ctx, Filter{"_id": r.InsertedID})
	e = r2.Decode(&result)
	return result, e
}

func (repo dbRepo[T]) UpdateById(ctx context.Context, id ID, updatedData T) (T, error) {
	var result T
	after := options.After
	r := repo.db.Collection(repo.collectionName).FindOneAndUpdate(ctx, &Filter{"id": id}, bson.M{
		"$set": updatedData,
	}, &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	})
	e := r.Decode(&result)
	return result, e
}

func (repo dbRepo[T]) DeleteById(ctx context.Context, id ID) error {
	var result T
	r := repo.db.Collection(repo.collectionName).FindOneAndDelete(ctx, &Filter{"id": id})
	e := r.Decode(&result)
	return e
}

func (repo dbRepo[T]) IndexFields(ctx context.Context) error {
	if repo.options == nil {
		return nil
	}
	models := make([]mongo.IndexModel, 0)
	for _, f := range repo.options.IndexFields {
		models = append(models, mongo.IndexModel{
			Keys: bson.D{{f, 1}},
		})
	}
	_, err := repo.db.Collection(repo.collectionName).Indexes().CreateMany(ctx, models)
	return err
}

//func (repo dbRepo[T]) Delete(ctx context.Context, query Query) error {
//	curr, err := repo.db.Collection(repo.collectionName).Find(ctx, query.filter, &options.FindOptions{
//		Sort: query.sort,
//	})
//	var res []T
//	curr.All(ctx, res)
//	for _, v := range res {
//		err = repo.DeleteById(ctx, v.GetId())
//		if err != nil {
//			return err
//		}
//	}
//	dr, e := repo.db.Collection(repo.collectionName).DeleteMany(ctx, query.filter)
//
//	return e
//}

type MongoRepoOptions struct {
	IndexFields []string
}

func NewMongoRepoAdapter[T Entity](
	db *mongo.Database,
	collectionName string,
	shortName string,
	o ...MongoRepoOptions,
) DbRepo[T] {
	if len(o) > 0 {
		return &dbRepo[T]{
			db,
			collectionName,
			shortName,
			&o[0],
		}
	}
	return &dbRepo[T]{
		db,
		collectionName,
		shortName,
		nil,
	}
}

func NewFxMongoRepo[T Entity](collectionName string, shortName string, indexFields []string) fx.Option {
	return fx.Module(
		"repo",
		fx.Provide(func(db *mongo.Database) DbRepo[T] {
			return NewMongoRepoAdapter[T](
				db,
				"devices",
				"dev",
				MongoRepoOptions{
					IndexFields: indexFields,
				},
			)
		}),
		fx.Invoke(func(lifecycle fx.Lifecycle, repo DbRepo[T]) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return repo.IndexFields(ctx)
				},
			})
		}),
	)
}
