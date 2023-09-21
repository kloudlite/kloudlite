package graph

import (
	"context"
	"fmt"
	"strings"

	"kloudlite.io/apps/container-registry/internal/domain"
	"kloudlite.io/common"
	"kloudlite.io/pkg/errors"
)

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
		err = errors.NewE(fmt.Errorf("%v", strings.Join(errMsgs, ",")))
	}

	return domain.RegistryContext{
		Context:     ctx,
		AccountName: accountName,

		UserId:    session.UserId,
		UserName:  session.UserName,
		UserEmail: session.UserEmail,
	}, err

}
