package graph

import (
	"context"
	"fmt"

	"github.com/kloudlite/api/common"
)

func GetUserSession(ctx context.Context) (*common.AuthSession, error) {
	session, ok := ctx.Value("user-session").(*common.AuthSession)
	if !ok {
		return nil, fmt.Errorf(`request context is missing 'user-session' key`)
	}
	return session, nil
}
