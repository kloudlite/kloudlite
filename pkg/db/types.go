package db

import (
	"go.mongodb.org/mongo-driver/bson"
)

type Record interface{}
type Opts map[string]interface {}
type Query bson.M

type Repo interface {
	Find(query Query, opts Opts) ([]Record, error)
	FindOne(query Query, opts Opts) (Record, error)
	Create(data Record)
}
