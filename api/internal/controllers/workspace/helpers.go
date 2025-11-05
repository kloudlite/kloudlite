package workspace

import (
	"context"
	"fmt"

	workmachinevl "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getWorkspaceTargetNamespace looks up the WorkMachine and returns its targetNamespace
// This is where the workspace's pods and services will be created
func (r *WorkspaceReconciler) getWorkspaceTargetNamespace(ctx context.Context, workspace *workspacev1.Workspace) (string, error) {
	if workspace.Spec.WorkmachineName == "" {
		return "", fmt.Errorf("workspace %s has no workmachineName set", workspace.Name)
	}

	// Fetch the WorkMachine
	workmachine := &workmachinevl.WorkMachine{}
	if err := r.Get(ctx, client.ObjectKey{Name: workspace.Spec.WorkmachineName}, workmachine); err != nil {
		return "", fmt.Errorf("failed to get workmachine %s: %w", workspace.Spec.WorkmachineName, err)
	}

	if workmachine.Spec.TargetNamespace == "" {
		return "", fmt.Errorf("workmachine %s has no targetNamespace set", workmachine.Name)
	}

	return workmachine.Spec.TargetNamespace, nil
}
