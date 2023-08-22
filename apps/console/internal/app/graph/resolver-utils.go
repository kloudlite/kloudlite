package graph

import (
	"context"
	"fmt"
	"kloudlite.io/apps/console/internal/app/graph/model"

	"kloudlite.io/apps/console/internal/domain"
)

func toConsoleContext(ctx context.Context) domain.ConsoleContext {
	if cc, ok := ctx.Value("kloudlite-ctx").(domain.ConsoleContext); ok {
		return cc
	}
	panic(fmt.Errorf("context values %q is missing", "kloudlite-ctx"))
}

func (r *queryResolver) getNamespaceFromProjectID(ctx context.Context, project model.ProjectID) (string, error) {
	switch project.Type {
	case model.ProjectIDTypeName:
		{
			proj, err := r.Domain.GetProject(toConsoleContext(ctx), project.Value)
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
			env, err := r.Domain.GetEnvironment(toConsoleContext(ctx), pTargetNs, scope.Value)
			if err != nil {
				return "", err
			}
			return env.Spec.TargetNamespace, nil
		}
	case model.WorkspaceOrEnvIDTypeWorkspaceName:
		{
			ws, err := r.Domain.GetWorkspace(toConsoleContext(ctx), pTargetNs, scope.Value)
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
