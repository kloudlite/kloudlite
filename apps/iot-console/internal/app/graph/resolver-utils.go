package graph

import (
	"context"
	"fmt"
	"github.com/kloudlite/api/apps/iot-console/internal/domain"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"strings"
)

func toIOTConsoleContext(ctx context.Context) (domain.IotConsoleContext, error) {
	missingContextValue := "context value (%s) is missing"

	errMsgs := make([]string, 0, 3)

	session, ok := ctx.Value("user-session").(*common.AuthSession)
	if !ok {
		errMsgs = append(errMsgs, fmt.Sprintf(missingContextValue, "user-session"))
	}

	accountName, ok := ctx.Value("account-name").(string)
	if !ok {
		errMsgs = append(errMsgs, fmt.Sprintf(missingContextValue, "account-name"))
	}

	var err error
	if len(errMsgs) != 0 {
		err = errors.NewE(errors.Newf("%v", strings.Join(errMsgs, ",")))
	}

	if err != nil {
		return domain.IotConsoleContext{}, errors.NewE(err)
	}

	return domain.IotConsoleContext{
		Context:     ctx,
		AccountName: accountName,

		UserId:    session.UserId,
		UserEmail: session.UserEmail,
		UserName:  session.UserName,
	}, nil
}

var (
	errNilDeployment      = errors.Newf("deployment object is nil")
	errNilProject         = errors.Newf("project object is nil")
	errNilEnvironment     = errors.Newf("environment object is nil")
	errNilDevice          = errors.Newf("device object is nil")
	errNilDeviceBlueprint = errors.Newf("device group object is nil")
	errNilApp             = errors.Newf("app object is nil")
)

func newIOTResourceContext(ctx domain.IotConsoleContext, projectName string, environmentName string) domain.IotResourceContext {
	return domain.IotResourceContext{
		IotConsoleContext: ctx,
		ProjectName:       projectName,
		EnvironmentName:   environmentName,
	}
}
