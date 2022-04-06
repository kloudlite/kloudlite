package mongo_db

import (
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewMongoDatabase(url string, dbName string) (db *mongo.Database, e error) {
	defer errors.HandleErr(&e)
	client, e := mongo.NewClient(options.Client().ApplyURI(url))
	errors.AssertNoError(e, fmt.Errorf("could not create mongo client"))
	return client.Database(dbName), nil
}
