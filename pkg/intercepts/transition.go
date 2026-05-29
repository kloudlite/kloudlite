package intercepts

import environmentv1 "github.com/kloudlite/kloudlite/types/environment/v1"

func UpsertConfig(intercepts []environmentv1.ServiceInterceptConfig, config environmentv1.ServiceInterceptConfig) ([]environmentv1.ServiceInterceptConfig, bool) {
	updated := make([]environmentv1.ServiceInterceptConfig, len(intercepts))
	copy(updated, intercepts)

	for i, existing := range updated {
		if existing.ServiceName == config.ServiceName {
			updated[i] = config
			return updated, true
		}
	}

	return append(updated, config), false
}

func RemoveService(intercepts []environmentv1.ServiceInterceptConfig, serviceName string) ([]environmentv1.ServiceInterceptConfig, bool) {
	updated := make([]environmentv1.ServiceInterceptConfig, 0, len(intercepts))
	removed := false
	for _, intercept := range intercepts {
		if intercept.ServiceName == serviceName {
			removed = true
			continue
		}
		updated = append(updated, intercept)
	}
	return updated, removed
}
