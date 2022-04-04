package repos

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/functions"
)

type dbRepo[T Entity] struct {
	db             *mongo.Database
	collectionName string
	shortName      string
}

var re = regexp.MustCompile("(\\W|_)+/g")

func (repo dbRepo[T]) NewId() ID {
	id, e := functions.CleanerNanoid(48)
	if e != nil {
		panic(fmt.Errorf("could not get cleanerNanoid()"))
	}
	return ID(fmt.Sprintf("%s-%s", repo.shortName, strings.ToLower(id)))
}

func (repo dbRepo[T]) Find(ctx context.Context, query Query) ([]T, error) {
	results := make([]T, 0)
	curr, err := repo.db.Collection(repo.collectionName).Find(ctx, query.filter, &options.FindOptions{
		Sort: query.sort,
	})
	err = curr.All(ctx, results)
	return results, err
}

func (repo dbRepo[T]) FindPaginated(ctx context.Context, query Query, page int64, size int64, opts ...Opts) (PaginatedRecord[T], error) {
	results := make([]T, 0)
	var offset int64 = (page - 1) * size
	curr, e := repo.db.Collection(repo.collectionName).Find(ctx, query.filter, &options.FindOptions{
		Limit: &size,
		Skip:  &offset,
		Sort:  query.sort,
	})
	e = curr.All(ctx, results)

	total, e := repo.db.Collection(repo.collectionName).CountDocuments(ctx, query.filter)

	return PaginatedRecord[T]{
		results:    results,
		totalCount: total,
	}, e
}

func (repo dbRepo[T]) FindById(ctx context.Context, id ID) (T, error) {
	var result T
	r := repo.db.Collection(repo.collectionName).FindOne(ctx, &Filter{"id": id})
	err := r.Decode(result)
	return result, err
}

func (repo dbRepo[T]) withId(data T) T {
	if data.GetId() != "" {
		return data
	}
	data, ok := data.SetId(repo.NewId()).(T)
	errors.Assert(ok, fmt.Errorf("could not typecast setId() into T"))
	return data
}

func (repo dbRepo[T]) Create(ctx context.Context, data T) (T, error) {
	var result T
	r, e := repo.db.Collection(repo.collectionName).InsertOne(ctx, repo.withId(data))
	fmt.Printf("%+v %+v", r, e)
	r2 := repo.db.Collection(repo.collectionName).FindOne(ctx, Filter{"_id": r.InsertedID})
	e = r2.Decode(&result)
	return result, e
}

func (repo dbRepo[T]) UpdateById(ctx context.Context, id ID, updatedData T) (T, error) {
	var result T
	r := repo.db.Collection(repo.collectionName).FindOneAndUpdate(ctx, &Filter{"id": id}, updatedData)
	e := r.Decode(&result)
	return result, e
}

func (repo dbRepo[T]) DeleteById(ctx context.Context, id ID) (T, error) {
	var result T
	r := repo.db.Collection(repo.collectionName).FindOneAndDelete(ctx, &Filter{"id": id})
	e := r.Decode(&result)
	return result, e
}

func (repo dbRepo[T]) Delete(ctx context.Context, query Query) error {
	_, e := repo.db.Collection(repo.collectionName).DeleteMany(ctx, query.filter)
	return e
}

func NewMongoRepoAdapter[T Entity](db *mongo.Database, collectionName string, shortName string) DbRepo[T] {
	return &dbRepo[T]{
		db,
		collectionName,
		shortName,
	}
}
