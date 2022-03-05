package db

import (
	"context"
	"fmt"

	"github.com/qiniu/qmgo"
	"kloudlite.io/pkg/errors"
)

func MakeMongoClient(ctx context.Context, uri string, db string) (cli *qmgo.QmgoClient, e error) {
	defer errors.HandleErr(&e)
	cli, e = qmgo.Open(ctx, &qmgo.Config{Uri: "mongodb://localhost:27017", Database: db})
	errors.AssertNoError(e, fmt.Errorf("failed to open mongo client"))
	return
}
