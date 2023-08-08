package domain

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"kloudlite.io/common"
	httpServer "kloudlite.io/pkg/http-server"
)

// access
const (
	READ_PROJECT   = "read_project"
	UPDATE_PROJECT = "update_project"

	READ_ACCOUNT   = "read_account"
	UPDATE_ACCOUNT = "update_account"
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

func GetUser(ctx AccountsContext) (string, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return "", errors.New("Unauthorized")
	}
	return string(session.UserId), nil
}
