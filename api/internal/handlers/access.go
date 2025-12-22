package handlers

import (
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
)

// HasRole checks if a user has a specific role
func HasRole(roles []platformv1alpha1.RoleType, targetRole platformv1alpha1.RoleType) bool {
	for _, r := range roles {
		if r == targetRole {
			return true
		}
	}
	return false
}

// UserHasAccessToEnvironment checks if a user has access to view an environment
func UserHasAccessToEnvironment(username string, env *environmentsv1.Environment) bool {
	// Owner always has access
	if env.Spec.OwnedBy == username {
		return true
	}

	visibility := env.Spec.Visibility
	if visibility == "" {
		visibility = "private"
	}

	switch visibility {
	case "private":
		return false
	case "shared":
		for _, sharedUser := range env.Spec.SharedWith {
			if sharedUser == username {
				return true
			}
		}
		return false
	case "open":
		return true
	default:
		return false
	}
}

// UserHasAccessToWorkspace checks if a user has access to view a workspace
func UserHasAccessToWorkspace(username string, ws *workspacesv1.Workspace) bool {
	// Owner always has access
	if ws.Spec.OwnedBy == username {
		return true
	}

	visibility := string(ws.Spec.Visibility)
	if visibility == "" {
		visibility = "private"
	}

	switch visibility {
	case "private":
		return false
	case "shared":
		for _, sharedUser := range ws.Spec.SharedWith {
			if sharedUser == username {
				return true
			}
		}
		return false
	case "open":
		return true
	default:
		return false
	}
}
