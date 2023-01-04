package util

import (
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func ReturnApp(app *entities.App) *model.App {

	services := make([]*model.ExposedService, 0)
	for _, s := range app.ExposedPorts {
		services = append(
			services, &model.ExposedService{
				Type:    string(s.Type),
				Target:  int(s.TargetPort),
				Exposed: int(s.Port),
			},
		)
	}

	return &model.App{
		IsFrozen:  app.Frozen,
		CreatedAt: app.CreationTime.String(),
		UpdatedAt: func() *string {
			if !app.UpdateTime.IsZero() {
				s := app.UpdateTime.String()
				return &s
			}
			return nil
		}(),
		IsLambda: app.IsLambda,
		Conditions: func() []*model.MetaCondition {
			conditions := make([]*model.MetaCondition, 0)
			for _, condition := range app.Conditions {
				conditions = append(
					conditions, &model.MetaCondition{
						Status:        string(condition.Status),
						ConditionType: condition.Type,
						LastTimeStamp: condition.LastTransitionTime.String(),
						Reason:        condition.Reason,
						Message:       condition.Message,
					},
				)
			}
			return conditions
		}(),
		ID:          app.Id,
		Name:        app.Name,
		Namespace:   app.Namespace,
		Description: app.Description,
		ReadableID:  repos.ID(app.ReadableId),
		Replicas:    &app.Replicas,
		AutoScale: func() *model.AutoScale {
			if app.AutoScale != nil {
				return &model.AutoScale{
					MinReplicas:     int(app.AutoScale.MinReplicas),
					MaxReplicas:     int(app.AutoScale.MaxReplicas),
					UsagePercentage: int(app.AutoScale.UsagePercentage),
				}
			}
			return nil
		}(),
		Services: services,
		Containers: func() []*model.AppContainer {
			containers := make([]*model.AppContainer, 0)
			for _, c := range app.Containers {
				envVars := make([]*model.EnvVar, 0)
				for _, e := range c.EnvVars {
					envVars = append(
						envVars, &model.EnvVar{
							Key: e.Key,
							Value: &model.EnvVal{
								Type:  e.Type,
								Value: e.Value,
								Ref:   e.Ref,
								Key:   e.RefKey,
							},
						},
					)
				}
				res := make([]*model.AttachedRes, 0)
				for _, r := range c.AttachedResources {
					res = append(
						res, &model.AttachedRes{
							ResID: r.ResourceId,
						},
					)
				}
				containers = append(
					containers, &model.AppContainer{
						Name:              c.Name,
						Image:             c.Image,
						PullSecret:        c.ImagePullSecret,
						EnvVars:           envVars,
						AttachedResources: res,
						ComputePlan:       c.ComputePlan,
						Quantity:          c.Quantity,
						IsShared:          &c.IsShared,
						Mounts: func() []*model.Mount {
							mounts := []*model.Mount{}
							for _, vm := range c.VolumeMounts {
								mounts = append(
									mounts, &model.Mount{
										Type: vm.Type,
										Ref:  vm.Ref,
										Path: vm.MountPath,
									},
								)
							}
							return mounts
						}(),
					},
				)
			}
			return containers
		}(),
		Project: &model.Project{ID: app.ProjectId},
		Status:  string(app.Status),
	}
}
