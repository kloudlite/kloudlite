package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/kloudlite/api/apps/container-registry/internal/domain"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func getUserId(ctx context.Context) (repos.ID, error) {
	session, ok := ctx.Value("user-session").(*common.AuthSession)

	if !ok {
		return "", errors.NewE(errors.Newf("context values %q is missing", "user-session"))
	}

	return session.UserId, nil
}

func toRegistryContext(ctx context.Context) (domain.RegistryContext, error) {
	session, ok := ctx.Value("user-session").(*common.AuthSession)

	errMsgs := []string{}
	if !ok {
		errMsgs = append(errMsgs, fmt.Sprintf("context values %q is missing", "user-session"))
	}

	accountName, ok := ctx.Value("account-name").(string)
	if !ok {
		errMsgs = append(errMsgs, fmt.Sprintf("context values %q is missing", "account-name"))
	}

	var err error
	if len(errMsgs) != 0 {
		err = errors.NewE(errors.Newf("%v", strings.Join(errMsgs, ",")))
	}

	return domain.RegistryContext{
		Context:     ctx,
		AccountName: accountName,

		UserId:    session.UserId,
		UserName:  session.UserName,
		UserEmail: session.UserEmail,
	}, errors.NewE(err)

}
