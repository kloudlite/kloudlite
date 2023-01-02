package graph

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/common"
	wErrors "kloudlite.io/pkg/errors"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

func returnApps(appEntities []*entities.App) []*model.App {
	apps := make([]*model.App, 0)
	for _, a := range appEntities {
		services := make([]*model.ExposedService, 0)
		for _, s := range a.ExposedPorts {
			services = append(
				services, &model.ExposedService{
					Type:    string(s.Type),
					Target:  int(s.TargetPort),
					Exposed: int(s.Port),
				},
			)
		}

		apps = append(
			apps, &model.App{
				IsFrozen:  a.Frozen,
				CreatedAt: a.CreationTime.String(),
				UpdatedAt: func() *string {
					if !a.UpdateTime.IsZero() {
						s := a.UpdateTime.String()
						return &s
					}
					return nil
				}(),
				IsLambda: a.IsLambda,
				Conditions: func() []*model.MetaCondition {
					conditions := make([]*model.MetaCondition, 0)
					for _, condition := range a.Conditions {
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
				ID:          a.Id,
				Name:        a.Name,
				Namespace:   a.Namespace,
				Description: a.Description,
				ReadableID:  repos.ID(a.ReadableId),
				AutoScale: func() *model.AutoScale {
					if a.AutoScale != nil {
						return &model.AutoScale{
							MinReplicas:     int(a.AutoScale.MinReplicas),
							MaxReplicas:     int(a.AutoScale.MaxReplicas),
							UsagePercentage: int(a.AutoScale.UsagePercentage),
						}
					}
					return nil
				}(),
				Replicas: &a.Replicas,
				Services: services,
				Containers: func() []*model.AppContainer {
					containers := make([]*model.AppContainer, 0)
					for _, c := range a.Containers {
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
				Project: &model.Project{ID: a.ProjectId},
				Status:  string(a.Status),
			},
		)
	}
	return apps
}

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

func withUserSession(ctx context.Context) (context.Context, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, wErrors.NotLoggedIn
	}
	return context.WithValue(ctx, "user_id", session.UserId), nil
}

func getInstances(d domain.Domain, obj *model.Environment, ctx context.Context, resType string) ([]*model.ResInstance, error) {

	if valid := d.ValidateResourecType(ctx, resType); !valid {
		return nil, fmt.Errorf(
			"resource type is not valid, use one of [%s, %s, %s, %s, %s, %s] resource type",
			common.ResourceApp,
			common.ResourceConfig,
			common.ResourceSecret,
			common.ResourceManagedService,
			common.ResourceManagedResource,
			common.ResourceRouter,
		)
	}

	ins, err := d.GetResInstances(ctx, obj.ID, resType)
	if err != nil {
		return nil, err
	}
	instances := make([]*model.ResInstance, 0)
	for _, instance := range ins {
		instances = append(instances, &model.ResInstance{
			ID:            instance.Id,
			ResourceID:    instance.ResourceId,
			EnvironmentID: instance.EnvironmentId,
			BlueprintID:   instance.BlueprintId,
			Overrides:     &instance.Overrides,
			ResourceType:  string(instance.ResourceType),
		})
	}

	var blueprintId *repos.ID
	// var parentEnvironmentId *repos.ID
	blueprintId = obj.BlueprintID
	// parentEnvironmentId = obj.ParentEnvironmentID

	if obj.BlueprintID == nil {
		o, err := d.GetEnvironment(ctx, obj.ID)
		if err != nil {
			return instances, nil
		}
		blueprintId = o.BlueprintId
		// parentEnvironmentId = o.ParentEnvironmentId
	}

	if blueprintId != nil {

		resIds := make([]repos.ID, 0)
		switch common.ResourceType(resType) {
		case common.ResourceApp:
			apps, err := d.GetApps(ctx, *obj.BlueprintID)
			if err != nil {
				break
			}
			fmt.Println(len(apps), obj.BlueprintID)

			for _, res := range apps {
				resIds = append(resIds, res.Id)
			}
			fmt.Println(resIds)

		case common.ResourceConfig:
			configs, err := d.GetConfigs(ctx, *obj.BlueprintID)
			if err != nil {
				break
			}
			for _, res := range configs {
				resIds = append(resIds, res.Id)
			}

		case common.ResourceSecret:
			secrets, err := d.GetSecrets(ctx, *obj.BlueprintID)
			if err != nil {
				break
			}
			for _, res := range secrets {
				resIds = append(resIds, res.Id)
			}

		case common.ResourceRouter:
			routers, err := d.GetRouters(ctx, *obj.BlueprintID)
			if err != nil {
				break
			}
			for _, res := range routers {
				resIds = append(resIds, res.Id)
			}

		case common.ResourceManagedService:
			svcs, err := d.GetManagedSvcs(ctx, *obj.BlueprintID)
			if err != nil {
				break
			}
			for _, res := range svcs {
				resIds = append(resIds, res.Id)
			}

		}

		// end of switch

		for _, res := range resIds {
			for _, instance := range instances {
				if res == instance.ID {
					continue
				}
			}

			ri, err := d.CreateResInstance(ctx, res, obj.ID, obj.BlueprintID, resType, "[]")
			if err != nil {
				fmt.Println("here",err)
				continue
			}

			instances = append(instances, &model.ResInstance{
				ID:            ri.Id,
				ResourceID:    ri.ResourceId,
				EnvironmentID: ri.EnvironmentId,
				BlueprintID:   ri.BlueprintId,
				Overrides:     &ri.Overrides,
				ResourceType:  resType,
			})
		}

	}

	return instances, nil
}
