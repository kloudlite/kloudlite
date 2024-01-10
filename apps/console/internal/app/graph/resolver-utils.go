package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/apps/console/internal/domain"
)

func toConsoleContext(ctx context.Context) (domain.ConsoleContext, error) {
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

	//clusterName, ok := ctx.Value("cluster-name").(string)
	//if !ok {
	//	errMsgs = append(errMsgs, fmt.Sprintf(missingContextValue, "cluster-name"))
	//}

	var err error
	if len(errMsgs) != 0 {
		err = errors.NewE(errors.Newf("%v", strings.Join(errMsgs, ",")))
	}

	if err != nil {
		return domain.ConsoleContext{}, errors.NewE(err)
	}

	return domain.ConsoleContext{
		Context:     ctx,
		AccountName: accountName,

		UserId:    session.UserId,
		UserEmail: session.UserEmail,
		UserName:  session.UserName,
	}, nil
}

// func (r *queryResolver) getNamespaceFromProjectAndScope(ctx context.Context, project model.ProjectID, scope model.WorkspaceOrEnvID) (string, error) {
// 	pTargetNs, err := r.getNamespaceFromProjectID(ctx, project)
// 	if err != nil {
// 		return "", errors.NewE(err)
// 	}
//
// 	switch scope.Type {
// 	case model.WorkspaceOrEnvIDTypeEnvironmentName:
// 		{
// 			cc, err := toConsoleContext(ctx)
// 			if err != nil {
// 				return "", errors.NewE(err)
// 			}
// 			env, err := r.Domain.GetEnvironment(cc, pTargetNs, scope.Value)
// 			if err != nil {
// 				return "", errors.NewE(err)
// 			}
// 			return env.Spec.TargetNamespace, nil
// 		}
// 	case model.WorkspaceOrEnvIDTypeWorkspaceName:
// 		{
// 			cc, err := toConsoleContext(ctx)
// 			if err != nil {
// 				return "", errors.NewE(err)
// 			}
// 			ws, err := r.Domain.GetEnvironment(cc, pTargetNs, scope.Value)
// 			if err != nil {
// 				return "", errors.NewE(err)
// 			}
// 			return ws.Spec.TargetNamespace, nil
// 		}
// 	case model.WorkspaceOrEnvIDTypeEnvironmentTargetNamespace:
// 		return scope.Value, nil
// 	case model.WorkspaceOrEnvIDTypeWorkspaceTargetNamespace:
// 		return scope.Value, nil
// 	default:
// 		return "", errors.Newf("invalid scope type %q", scope.Type)
// 	}
// }

var (
	errNilApp                   = errors.Newf("app obj is nil")
	errNilConfig                = errors.Newf("config obj is nil")
	errNilSecret                = errors.Newf("secret obj is nil")
	errNilEnvironment           = errors.Newf("environment obj is nil")
	errNilVPNDevice             = errors.Newf("vpn device obj is nil")
	errNilImagePullSecret       = errors.Newf("imagePullSecret obj is nil")
	errNilManagedResource       = errors.Newf("managed resource obj is nil")
	errNilProject               = errors.Newf("project obj is nil")
	errNilProjectManagedService = errors.Newf("project manged svc obj is nil")
	errNilRouter                = errors.Newf("router obj is nil")
)

func newResourceContext(ctx domain.ConsoleContext, projectName string, environmentName string) domain.ResourceContext {
	return domain.ResourceContext{
		ConsoleContext:  ctx,
		ProjectName:     projectName,
		EnvironmentName: environmentName,
	}
}
