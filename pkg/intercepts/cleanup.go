package intercepts

import environmentv1 "github.com/kloudlite/kloudlite/types/environment/v1"

func RemoveForWorkspace(intercepts []environmentv1.ServiceInterceptConfig, workspaceName, workspaceNamespace string) ([]environmentv1.ServiceInterceptConfig, []string) {
	if intercepts == nil {
		return nil, nil
	}

	kept := make([]environmentv1.ServiceInterceptConfig, 0, len(intercepts))
	removed := []string{}
	for _, intercept := range intercepts {
		if intercept.WorkspaceRef != nil &&
			intercept.WorkspaceRef.Name == workspaceName &&
			intercept.WorkspaceRef.Namespace == workspaceNamespace {
			removed = append(removed, intercept.ServiceName)
			continue
		}
		kept = append(kept, intercept)
	}

	return kept, removed
}
