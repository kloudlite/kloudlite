package intercepts

import environmentv1 "github.com/kloudlite/kloudlite/types/environment/v1"

type ActiveStatus struct {
	EnvironmentName string
	ServiceName     string
	Phase           string
	Message         string
}

type ContextPortMapping struct {
	ServicePort   int32 `json:"servicePort"`
	WorkspacePort int32 `json:"workspacePort"`
}

type ContextEntry struct {
	ServiceName  string               `json:"serviceName"`
	PortMappings []ContextPortMapping `json:"portMappings"`
}

func ActiveStatuses(env *environmentv1.Environment, workspaceName string) []ActiveStatus {
	if env == nil || env.Status.ComposeStatus == nil {
		return nil
	}

	statuses := []ActiveStatus{}
	for _, status := range env.Status.ComposeStatus.ActiveIntercepts {
		if workspaceName != "" && status.WorkspaceName != workspaceName {
			continue
		}
		statuses = append(statuses, ActiveStatus{
			EnvironmentName: env.Name,
			ServiceName:     status.ServiceName,
			Phase:           status.Phase,
			Message:         status.Message,
		})
	}
	return statuses
}

func ContextForWorkspace(env *environmentv1.Environment, workspaceName string) []ContextEntry {
	if env == nil || env.Status.ComposeStatus == nil {
		return []ContextEntry{}
	}

	entries := []ContextEntry{}
	for _, status := range env.Status.ComposeStatus.ActiveIntercepts {
		if status.WorkspaceName != workspaceName {
			continue
		}
		entries = append(entries, ContextEntry{
			ServiceName:  status.ServiceName,
			PortMappings: contextPortMappings(env, status.ServiceName),
		})
	}
	return entries
}

func contextPortMappings(env *environmentv1.Environment, serviceName string) []ContextPortMapping {
	if env.Spec.Compose == nil {
		return nil
	}

	for _, intercept := range env.Spec.Compose.Intercepts {
		if intercept.ServiceName != serviceName {
			continue
		}

		mappings := make([]ContextPortMapping, 0, len(intercept.PortMappings))
		for _, mapping := range intercept.PortMappings {
			mappings = append(mappings, ContextPortMapping{
				ServicePort:   mapping.ServicePort,
				WorkspacePort: mapping.WorkspacePort,
			})
		}
		return mappings
	}

	return nil
}
