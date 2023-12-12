package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/kloudlite/api/apps/console/internal/app/graph/model"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/apps/console/internal/domain"
)

func toConsoleContext(ctx context.Context) (domain.ConsoleContext, error) {
	session, ok := ctx.Value("user-session").(*common.AuthSession)

	errMsgs := []string{}
	if !ok {
		errMsgs = append(errMsgs, fmt.Sprintf("context values %q is missing", "user-session"))
	}

	accountName, ok := ctx.Value("account-name").(string)
	if !ok {
		errMsgs = append(errMsgs, fmt.Sprintf("context values %q is missing", "account-name"))
	}

	clusterName, ok := ctx.Value("cluster-name").(string)
	if !ok {
		errMsgs = append(errMsgs, fmt.Sprintf("context values %q is missing", "cluster-name"))
	}

	var err error
	if len(errMsgs) != 0 {
		err = errors.NewE(fmt.Errorf("%v", strings.Join(errMsgs, ",")))
	}

	return domain.ConsoleContext{
		Context:     ctx,
		ClusterName: clusterName,
		AccountName: accountName,

		UserId:    session.UserId,
		UserEmail: session.UserEmail,
		UserName:  session.UserName,
	}, err
}

func (r *queryResolver) getNamespaceFromProjectID(ctx context.Context, project model.ProjectID) (string, error) {
	switch project.Type {
	case model.ProjectIDTypeName:
		{
			cc, err := toConsoleContext(ctx)
			if err != nil {
				return "", err
			}
			proj, err := r.Domain.GetProject(cc, project.Value)
			if err != nil {
				return "", err
			}
			return proj.Spec.TargetNamespace, nil
		}
	case model.ProjectIDTypeTargetNamespace:
		{
			return project.Value, nil
		}
	default:
		return "", fmt.Errorf("invalid project type %q", project.Type)
	}
}

func (r *queryResolver) getNamespaceFromProjectAndScope(ctx context.Context, project model.ProjectID, scope model.WorkspaceOrEnvID) (string, error) {
	pTargetNs, err := r.getNamespaceFromProjectID(ctx, project)
	if err != nil {
		return "", err
	}

	switch scope.Type {
	case model.WorkspaceOrEnvIDTypeEnvironmentName:
		{
			cc, err := toConsoleContext(ctx)
			if err != nil {
				return "", err
			}
			env, err := r.Domain.GetEnvironment(cc, pTargetNs, scope.Value)
			if err != nil {
				return "", err
			}
			return env.Spec.TargetNamespace, nil
		}
	case model.WorkspaceOrEnvIDTypeWorkspaceName:
		{
			cc, err := toConsoleContext(ctx)
			if err != nil {
				return "", err
			}
			ws, err := r.Domain.GetWorkspace(cc, pTargetNs, scope.Value)
			if err != nil {
				return "", err
			}
			return ws.Spec.TargetNamespace, nil
		}
	case model.WorkspaceOrEnvIDTypeEnvironmentTargetNamespace:
		return scope.Value, nil
	case model.WorkspaceOrEnvIDTypeWorkspaceTargetNamespace:
		return scope.Value, nil
	default:
		return "", fmt.Errorf("invalid scope type %q", scope.Type)
	}
}
