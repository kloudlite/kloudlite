package graph

import (
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain/entities"
	"strconv"
)

func projectModelFromEntity(projectEntity *entities.Project) *model.Project {
	return &model.Project{
		ID:          projectEntity.Id,
		Name:        projectEntity.Name,
		DisplayName: projectEntity.DisplayName,
		ReadableID:  projectEntity.ReadableId,
		Logo:        projectEntity.Logo,
		Description: projectEntity.Description,
		Account: &model.Account{
			ID: projectEntity.AccountId,
		},
	}
}

func configModelFromEntity(configEntity *entities.Config) *model.Config {
	entries := make([]*model.CSEntry, 0)
	for _, e := range configEntity.Data {
		entries = append(entries, &model.CSEntry{
			Key:   e.Key,
			Value: e.Value,
		})
	}
	return &model.Config{
		ID:      configEntity.Id,
		Name:    configEntity.Name,
		Project: &model.Project{ID: configEntity.ProjectId},
		Entries: entries,
	}
}

func routerModelFromEntity(routerEntity *entities.Router) *model.Router {
	entries := make([]*model.Route, 0)
	for _, e := range routerEntity.Routes {
		atoi, _ := strconv.Atoi(e.Port)
		// TODO: handle error
		entries = append(entries, &model.Route{
			Path:    e.Path,
			AppName: e.AppName,
			Port:    atoi,
		})
	}
	d := routerEntity.Domains
	if d == nil {
		d = []string{}
	}
	return &model.Router{
		ID:      routerEntity.Id,
		Name:    routerEntity.Name,
		Project: &model.Project{ID: routerEntity.ProjectId},
		Domains: d,
		Routes:  entries,
	}
}

func secretModelFromEntity(secretEntity *entities.Secret) *model.Secret {
	entries := make([]*model.CSEntry, 0)
	for _, e := range secretEntity.Data {
		entries = append(entries, &model.CSEntry{
			Key:   e.Key,
			Value: e.Value,
		})
	}
	return &model.Secret{
		ID:      secretEntity.Id,
		Name:    secretEntity.Name,
		Project: &model.Project{ID: secretEntity.ProjectId},
		Entries: entries,
	}
}

func managedSvcModelFromEntity(svcEntity *entities.ManagedService) *model.ManagedSvc {
	return &model.ManagedSvc{
		ID:      svcEntity.Id,
		Name:    svcEntity.Name,
		Project: &model.Project{ID: svcEntity.ProjectId},
		Source:  string(svcEntity.ServiceType),
		Values:  svcEntity.Values,
	}
}

func managedResourceModelFromEntity(resEntity *entities.ManagedResource) *model.ManagedRes {
	return &model.ManagedRes{
		ID:           resEntity.Id,
		Name:         resEntity.Name,
		ResourceType: string(resEntity.ResourceType),
		Installation: &model.ManagedSvc{
			ID: resEntity.ServiceId,
		},
		Values: resEntity.Values,
	}
}
