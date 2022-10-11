package domain

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"kloudlite.io/common"
	httpServer "kloudlite.io/pkg/http-server"
)

const (
	ReadProject   = "read_project"
	UpdateProject = "update_project"
	ReadAccount   = "read_account"
)

func mongoError(err error, descp string) error {
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New(descp)
		}
		return err
	}
	return nil
}

func GetUser(ctx context.Context) (string, error) {

	session := httpServer.GetSession[*common.AuthSession](ctx)

	if session == nil {
		return "", errors.New("Unauthorized")
	}
	return string(session.UserId), nil
}
