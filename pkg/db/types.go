package db

import (
	"go.mongodb.org/mongo-driver/bson"
)

type Record interface{}
type Query bson.M

type MongoCollection interface {
	Find(query interface{})
	Create(query bson.M) Record
}
