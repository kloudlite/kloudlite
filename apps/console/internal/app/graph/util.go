package graph

import (
	"context"
	"fmt"
	"strings"

	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/common"
	wErrors "kloudlite.io/pkg/errors"
	fn "kloudlite.io/pkg/functions"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

func NewId(shortName string) repos.ID {
	id, e := fn.CleanerNanoid(28)
	if e != nil {
		panic(fmt.Errorf("could not get cleanerNanoid()"))
	}
	return repos.ID(fmt.Sprintf("%s-%s", shortName, strings.ToLower(id)))
}

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

func withUserSession(ctx context.Context) (context.Context, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, wErrors.NotLoggedIn
	}
	return context.WithValue(ctx, "user_id", session.UserId), nil
}

func getInstances(d domain.Domain, obj *model.Environment, ctx context.Context, resType string) ([]*model.ResInstance, error) {

	if err := d.ValidateResourecType(ctx, resType); err != nil {
		return nil, err
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
			Enabled:       instance.Enabled,
			ResourceType:  string(instance.ResourceType),
		})
	}

	return instances, nil
}
