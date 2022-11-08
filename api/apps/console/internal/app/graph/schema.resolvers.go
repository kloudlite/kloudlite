package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/apps/console/internal/domain/entities/localenv"
	"kloudlite.io/common"
	wErrors "kloudlite.io/pkg/errors"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

func (r *accountResolver) Projects(ctx context.Context, obj *model.Account) ([]*model.Project, error) {
	projectEntities, err := r.Domain.GetAccountProjects(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	projectModels := make([]*model.Project, 0)
	for _, pe := range projectEntities {
		projectModels = append(projectModels, projectModelFromEntity(pe))
	}
	return projectModels, err
}

func (r *accountResolver) Devices(ctx context.Context, obj *model.Account) ([]*model.Device, error) {
	deviceEntities, e := r.Domain.ListAccountDevices(ctx, obj.ID)
	wErrors.AssertNoError(e, fmt.Errorf("not able to list devices of account %s", obj.ID))
	devices := make([]*model.Device, 0)
	for _, device := range deviceEntities {
		devices = append(
			devices, &model.Device{
				ID:      device.Id,
				User:    &model.User{ID: device.UserId},
				Account: &model.Account{ID: device.AccountId},
				Name:    device.Name,
				Region:  device.ActiveRegion,
				Ports: func() []*model.Port {
					var ports []*model.Port
					for _, port := range device.ExposedPorts {
						ports = append(
							ports, &model.Port{
								Port: int(port.Port),
								TargetPort: func() *int {
									if port.TargetPort != nil {
										i := int(*port.TargetPort)
										return &i
									}
									return nil
								}(),
							},
						)
					}
					return ports
				}(),
			},
		)
	}
	return devices, e
}

func (r *appResolver) Restart(ctx context.Context, obj *model.App) (bool, error) {
	err := r.Domain.RestartApp(ctx, obj.ID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *appResolver) DoFreeze(ctx context.Context, obj *model.App) (bool, error) {
	err := r.Domain.FreezeApp(ctx, obj.ID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *appResolver) DoUnfreeze(ctx context.Context, obj *model.App) (bool, error) {
	err := r.Domain.UnFreezeApp(ctx, obj.ID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *appResolver) Intercept(ctx context.Context, obj *model.App, deviceID repos.ID) (bool, error) {
	err := r.Domain.InterceptApp(ctx, obj.ID, deviceID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *appResolver) CloseIntercept(ctx context.Context, obj *model.App) (bool, error) {
	err := r.Domain.CloseIntercept(ctx, obj.ID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *cloudProviderResolver) Edges(ctx context.Context, obj *model.CloudProvider) ([]*model.EdgeRegion, error) {
	regions, err := r.Domain.GetEdgeRegions(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	res := make([]*model.EdgeRegion, 0)
	for _, r := range regions {
		res = append(
			res, &model.EdgeRegion{
				ID:        r.Id,
				Name:      r.Name,
				Region:    r.Region,
				CreatedAt: r.CreationTime.String(),
				UpdatedAt: func() *string {
					if !r.UpdateTime.IsZero() {
						s := r.UpdateTime.String()
						return &s
					}
					return nil
				}(),
				Pools: func() []*model.NodePool {
					pools := make([]*model.NodePool, 0)
					for _, pool := range r.Pools {
						pools = append(
							pools, &model.NodePool{
								Name:   pool.Name,
								Config: pool.Config,
								Min:    pool.Min,
								Max:    pool.Max,
							},
						)
					}
					return pools
				}(),
			},
		)
	}
	return res, nil
}

func (r *deviceResolver) User(ctx context.Context, obj *model.Device) (*model.User, error) {
	var e error
	defer wErrors.HandleErr(&e)
	device := obj
	deviceEntity, e := r.Domain.GetDevice(ctx, device.ID)
	wErrors.AssertNoError(e, fmt.Errorf("not able to get device"))
	return &model.User{
		ID: deviceEntity.UserId,
	}, e
}

func (r *deviceResolver) Configuration(ctx context.Context, obj *model.Device) (map[string]interface{}, error) {
	return r.Domain.GetDeviceConfig(ctx, obj.ID)
}

func (r *deviceResolver) Account(ctx context.Context, obj *model.Device) (*model.Account, error) {
	var e error
	defer wErrors.HandleErr(&e)
	device := obj
	deviceEntity, e := r.Domain.GetDevice(ctx, device.ID)
	wErrors.AssertNoError(e, fmt.Errorf("not able to get device"))
	return &model.Account{
		ID: deviceEntity.AccountId,
	}, e
}

func (r *deviceResolver) InterceptingServices(ctx context.Context, obj *model.Device) ([]*model.App, error) {
	appEntities, err := r.Domain.GetInterceptedApps(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	apps := returnApps(appEntities)
	return apps, err
}

func (r *edgeRegionResolver) Provider(ctx context.Context, obj *model.EdgeRegion) (*model.CloudProvider, error) {
	region, err := r.Domain.GetEdgeRegion(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	provider, err := r.Domain.GetCloudProvider(ctx, region.ProviderId)
	if err != nil {
		return nil, err
	}
	return &model.CloudProvider{
		ID:       provider.Id,
		Name:     provider.Name,
		Provider: provider.Provider,
		IsShared: *provider.AccountId == "kl-core",
	}, nil
}

func (r *managedResResolver) Outputs(ctx context.Context, obj *model.ManagedRes) (map[string]interface{}, error) {
	if strings.HasPrefix(string(obj.ID), "mgsvc-") {
		output, err := r.Domain.GetManagedSvcOutput(ctx, obj.ID)
		if err != nil {
			return nil, err
		}
		return output, nil
	}

	outputs, err := r.Domain.GetManagedResOutput(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	return outputs, nil
}

func (r *managedSvcResolver) Resources(ctx context.Context, obj *model.ManagedSvc) ([]*model.ManagedRes, error) {
	resources, err := r.Domain.GetManagedResourcesOfService(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	res := make([]*model.ManagedRes, 0)
	for _, r := range resources {
		res = append(res, managedResourceModelFromEntity(r))
	}
	return res, nil
}

func (r *managedSvcResolver) Outputs(ctx context.Context, obj *model.ManagedSvc) (map[string]interface{}, error) {
	output, err := r.Domain.GetManagedSvcOutput(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (r *mutationResolver) MangedSvcInstall(ctx context.Context, projectID repos.ID, category repos.ID, serviceType repos.ID, name string, values map[string]interface{}) (*model.ManagedSvc, error) {
	svcEntity, err := r.Domain.InstallManagedSvc(ctx, projectID, serviceType, name, values)
	if err != nil {
		return nil, err
	}
	return &model.ManagedSvc{
		ID:      svcEntity.Id,
		Name:    svcEntity.Name,
		Project: &model.Project{ID: projectID},
		Source:  string(svcEntity.ServiceType),
		Values:  values,
	}, nil
}

func (r *mutationResolver) MangedSvcUninstall(ctx context.Context, installationID repos.ID) (bool, error) {
	uninstall, err := r.Domain.UnInstallManagedSvc(ctx, installationID)
	if err != nil {
		return false, err
	}
	return uninstall, nil
}

func (r *mutationResolver) MangedSvcUpdate(ctx context.Context, installationID repos.ID, values map[string]interface{}) (bool, error) {
	svc, err := r.Domain.UpdateManagedSvc(ctx, installationID, values)
	if err != nil {
		return false, err
	}
	return svc, nil
}

func (r *mutationResolver) ManagedResCreate(ctx context.Context, installationID repos.ID, name string, resourceType string, values map[string]interface{}) (*model.ManagedRes, error) {
	kvs := make(map[string]string, 0)
	for k, v := range values {
		kvs[k] = v.(string)
	}
	res, err := r.Domain.InstallManagedRes(ctx, installationID, name, resourceType, kvs)
	if err != nil {
		return nil, err
	}
	return &model.ManagedRes{
		ID:           res.Id,
		Name:         res.Name,
		ResourceType: string(res.ResourceType),
		Installation: &model.ManagedSvc{
			ID: installationID,
		},
		Values: values,
	}, nil
}

func (r *mutationResolver) ManagedResUpdate(ctx context.Context, resID repos.ID, values map[string]interface{}) (bool, error) {
	res, err := r.Domain.UpdateManagedRes(
		ctx, resID, func() map[string]string {
			val := make(map[string]string, 0)
			for k, v := range values {
				values[k] = v.(string)
			}
			return val
		}(),
	)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (r *mutationResolver) ManagedResDelete(ctx context.Context, resID repos.ID) (*bool, error) {
	res, err := r.Domain.UnInstallManagedRes(ctx, resID)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *mutationResolver) CoreAddDevice(ctx context.Context, accountID repos.ID, name string) (*model.Device, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("user not logged in")
	}
	ctx = context.WithValue(ctx, "user_id", session.UserId)
	device, err := r.Domain.AddDevice(ctx, name, accountID, session.UserId)
	if err != nil {
		return nil, err
	}
	return &model.Device{
		ID: device.Id,
		User: &model.User{
			ID: session.UserId,
		},
		Region: device.ActiveRegion,
		Name:   device.Name,
		Account: &model.Account{
			ID: device.AccountId,
		},
		Ports: func() []*model.Port {
			var ports []*model.Port
			for _, port := range device.ExposedPorts {
				ports = append(
					ports, &model.Port{
						Port: int(port.Port),
						TargetPort: func() *int {
							if port.TargetPort != nil {
								i := int(*port.TargetPort)
								return &i
							}
							return nil
						}(),
					},
				)
			}
			return ports
		}(),
	}, nil
}

func (r *mutationResolver) CoreRemoveDevice(ctx context.Context, deviceID repos.ID) (bool, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("user not logged in")
	}
	err := r.Domain.RemoveDevice(ctx, deviceID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreUpdateDevice(ctx context.Context, deviceID repos.ID, name *string, region *string, ports []*model.PortIn) (bool, error) {
	_, err := r.Domain.UpdateDevice(
		ctx, deviceID, name, region, func() []entities.Port {
			makePorts := make([]entities.Port, 0)
			for _, p := range ports {
				makePorts = append(
					makePorts, entities.Port{
						Port: int32(p.Port),
						TargetPort: func() *int32 {
							if p.TargetPort != nil {
								i := int32(*p.TargetPort)
								return &i
							}
							return nil
						}(),
					},
				)
			}
			return makePorts
		}(),
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreCreateProject(ctx context.Context, accountID repos.ID, name string, displayName string, logo *string, description *string, regionID *repos.ID) (*model.Project, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("user not logged in")
	}
	projectEntity, err := r.Domain.CreateProject(ctx, session.UserId, accountID, name, displayName, logo, regionID, description)
	if err != nil {
		return nil, err
	}
	return projectModelFromEntity(projectEntity), nil
}

func (r *mutationResolver) CoreUpdateProject(ctx context.Context, projectID repos.ID, displayName *string, cluster *string, logo *string, description *string) (bool, error) {
	return r.Domain.UpdateProject(ctx, projectID, displayName, cluster, logo, description)
}

func (r *mutationResolver) CoreDeleteProject(ctx context.Context, projectID repos.ID) (bool, error) {
	return r.Domain.DeleteProject(ctx, projectID)
}

func (r *mutationResolver) IamInviteProjectMember(ctx context.Context, projectID repos.ID, email string, role string) (bool, error) {
	return r.Domain.InviteProjectMember(ctx, projectID, email, role)
}

func (r *mutationResolver) IamRemoveProjectMember(ctx context.Context, projectID repos.ID, userID repos.ID) (bool, error) {
	err := r.Domain.RemoveProjectMember(ctx, projectID, userID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) IamUpdateProjectMember(ctx context.Context, projectID repos.ID, userID repos.ID, role string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateApp(ctx context.Context, projectID repos.ID, app model.AppInput) (*model.App, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("user not logged in")
	}

	ports := make([]entities.ExposedPort, 0)
	for _, port := range app.Services {
		ports = append(
			ports, entities.ExposedPort{
				Port:       int64(port.Exposed),
				TargetPort: int64(port.Target),
				Type:       entities.PortType(port.Type),
			},
		)
	}
	containers := make([]entities.Container, 0)
	for _, container := range app.Containers {
		e := make([]entities.EnvVar, 0)
		for _, env := range container.EnvVars {
			e = append(
				e, entities.EnvVar{
					Key:    env.Key,
					Type:   env.Value.Type,
					Value:  env.Value.Value,
					Ref:    env.Value.Ref,
					RefKey: env.Value.Key,
				},
			)
		}
		a := make([]entities.AttachedResource, 0)
		for _, attached := range container.AttachedResources {
			a = append(
				a, entities.AttachedResource{
					ResourceId: attached.ResID,
				},
			)
		}

		in := entities.Container{
			Name:  container.Name,
			Image: container.Image,
			IsShared: func() bool {
				if container.IsShared != nil {
					return *container.IsShared
				}
				return false
			}(),
			VolumeMounts: func() []entities.VolumeMount {
				if container.Mounts == nil {
					return nil
				}
				if len(container.Mounts) == 0 {
					return nil
				}
				out := make([]entities.VolumeMount, 0)
				for _, mount := range container.Mounts {
					out = append(
						out, entities.VolumeMount{
							MountPath: mount.Path,
							Type:      mount.Type,
							Ref:       mount.Ref,
						},
					)
				}
				return out
			}(),
			ImagePullSecret:   container.PullSecret,
			EnvVars:           e,
			ComputePlan:       container.ComputePlan,
			Quantity:          container.Quantity,
			AttachedResources: a,
		}
		containers = append(containers, in)
	}
	entity, err := r.Domain.InstallApp(
		ctx, projectID, entities.App{
			Name:      app.Name,
			IsLambda:  app.IsLambda,
			ProjectId: projectID,
			AutoScale: func() *entities.AutoScale {
				if app.AutoScale != nil {
					return &entities.AutoScale{
						MinReplicas:     int64(app.AutoScale.MinReplicas),
						MaxReplicas:     int64(app.AutoScale.MaxReplicas),
						UsagePercentage: int64(app.AutoScale.UsagePercentage),
					}
				}
				return nil
			}(),
			ReadableId:   string(app.ReadableID),
			Description:  app.Description,
			Replicas:     1,
			ExposedPorts: ports,
			Containers:   containers,
			Metadata: func() map[string]any {
				return app.Metadata
			}(),
		},
	)
	if err != nil {
		return nil, err
	}
	return &model.App{
		CreatedAt: entity.CreationTime.String(),
		UpdatedAt: func() *string {
			if !entity.UpdateTime.IsZero() {
				s := entity.UpdateTime.String()
				return &s
			}
			return nil
		}(),
		Conditions: func() []*model.MetaCondition {
			conditions := make([]*model.MetaCondition, 0)
			for _, condition := range entity.Conditions {
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
		ID:          entity.Id,
		Name:        entity.Name,
		Namespace:   entity.Namespace,
		IsLambda:    entity.IsLambda,
		Description: entity.Description,
		ReadableID:  repos.ID(entity.ReadableId),
		Replicas:    &entity.Replicas,
		IsFrozen:    entity.Frozen,
		AutoScale: func() *model.AutoScale {
			if entity.AutoScale != nil {
				return &model.AutoScale{
					MinReplicas:     int(entity.AutoScale.MinReplicas),
					MaxReplicas:     int(entity.AutoScale.MaxReplicas),
					UsagePercentage: int(entity.AutoScale.UsagePercentage),
				}
			}
			return nil
		}(),
		Services: func() []*model.ExposedService {
			services := make([]*model.ExposedService, 0)
			for _, port := range entity.ExposedPorts {
				services = append(
					services, &model.ExposedService{
						Exposed: int(port.Port),
						Target:  int(port.TargetPort),
						Type:    string(port.Type),
					},
				)
			}
			return services
		}(),
		Containers: func() []*model.AppContainer {
			containers := make([]*model.AppContainer, 0)
			for _, container := range entity.Containers {
				c := &model.AppContainer{
					Name:        container.Name,
					Image:       container.Image,
					PullSecret:  container.ImagePullSecret,
					ComputePlan: container.ComputePlan,
					Quantity:    container.Quantity,
					AttachedResources: func() []*model.AttachedRes {
						attached := make([]*model.AttachedRes, 0)
						for _, attachedResource := range container.AttachedResources {
							attached = append(
								attached, &model.AttachedRes{
									ResID: attachedResource.ResourceId,
								},
							)
						}
						return attached
					}(),
					EnvVars: func() []*model.EnvVar {
						envVars := make([]*model.EnvVar, 0)
						for _, envVar := range container.EnvVars {
							envVars = append(
								envVars, &model.EnvVar{
									Key: envVar.Key,
									Value: &model.EnvVal{
										Type:  envVar.Type,
										Value: envVar.Value,
									},
								},
							)
						}
						return envVars
					}(),
				}
				containers = append(containers, c)
			}
			return containers
		}(),
	}, nil
}

func (r *mutationResolver) CoreUpdateApp(ctx context.Context, projectID repos.ID, appID repos.ID, app model.AppInput) (*model.App, error) {
	session := httpServer.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("user not logged in")
	}
	ports := make([]entities.ExposedPort, 0)
	for _, port := range app.Services {
		ports = append(
			ports, entities.ExposedPort{
				Port:       int64(port.Exposed),
				TargetPort: int64(port.Target),
				Type:       entities.PortType(port.Type),
			},
		)
	}
	containers := make([]entities.Container, 0)
	for _, container := range app.Containers {
		e := make([]entities.EnvVar, 0)
		for _, env := range container.EnvVars {
			e = append(
				e, entities.EnvVar{
					Key:    env.Key,
					Type:   env.Value.Type,
					Value:  env.Value.Value,
					Ref:    env.Value.Ref,
					RefKey: env.Value.Key,
				},
			)
		}
		a := make([]entities.AttachedResource, 0)
		for _, attached := range container.AttachedResources {
			a = append(
				a, entities.AttachedResource{
					ResourceId: attached.ResID,
				},
			)
		}

		in := entities.Container{
			Name:  container.Name,
			Image: container.Image,
			IsShared: func() bool {
				if container.IsShared != nil {
					return *container.IsShared
				}
				return false
			}(),
			VolumeMounts: func() []entities.VolumeMount {
				if container.Mounts == nil {
					return nil
				}
				if len(container.Mounts) == 0 {
					return nil
				}
				out := make([]entities.VolumeMount, 0)
				for _, mount := range container.Mounts {
					out = append(
						out, entities.VolumeMount{
							MountPath: mount.Path,
							Type:      mount.Type,
							Ref:       mount.Ref,
						},
					)
				}
				return out
			}(),
			ImagePullSecret:   container.PullSecret,
			EnvVars:           e,
			ComputePlan:       container.ComputePlan,
			Quantity:          container.Quantity,
			AttachedResources: a,
		}
		containers = append(containers, in)
	}
	entity, err := r.Domain.UpdateApp(
		ctx, appID, entities.App{
			Name:        app.Name,
			ProjectId:   projectID,
			IsLambda:    app.IsLambda,
			ReadableId:  string(app.ReadableID),
			Description: app.Description,
			Replicas: func() int {
				if app.Replicas != nil {
					return *app.Replicas
				}
				return 1
			}(),
			AutoScale: func() *entities.AutoScale {
				if app.AutoScale != nil {
					return &entities.AutoScale{
						MinReplicas:     int64(app.AutoScale.MinReplicas),
						MaxReplicas:     int64(app.AutoScale.MaxReplicas),
						UsagePercentage: int64(app.AutoScale.UsagePercentage),
					}
				}
				return nil
			}(),
			ExposedPorts: ports,
			Containers:   containers,
			Metadata: func() map[string]any {
				return app.Metadata
			}(),
		},
	)
	if err != nil {
		return nil, err
	}

	return &model.App{
		CreatedAt: entity.CreationTime.String(),
		UpdatedAt: func() *string {
			if !entity.UpdateTime.IsZero() {
				s := entity.UpdateTime.String()
				return &s
			}
			return nil
		}(),
		IsFrozen: entity.Frozen,
		Conditions: func() []*model.MetaCondition {
			conditions := make([]*model.MetaCondition, 0)
			for _, condition := range entity.Conditions {
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
		ID:          entity.Id,
		Name:        entity.Name,
		IsLambda:    entity.IsLambda,
		Namespace:   entity.Namespace,
		Description: entity.Description,
		ReadableID:  repos.ID(entity.ReadableId),
		Replicas:    &entity.Replicas,
		AutoScale: func() *model.AutoScale {
			if entity.AutoScale != nil {
				return &model.AutoScale{
					MinReplicas:     int(entity.AutoScale.MinReplicas),
					MaxReplicas:     int(entity.AutoScale.MaxReplicas),
					UsagePercentage: int(entity.AutoScale.UsagePercentage),
				}
			}
			return nil
		}(),
		Services: func() []*model.ExposedService {
			services := make([]*model.ExposedService, 0)
			for _, port := range entity.ExposedPorts {
				services = append(
					services, &model.ExposedService{
						Exposed: int(port.Port),
						Target:  int(port.TargetPort),
						Type:    string(port.Type),
					},
				)
			}
			return services
		}(),
		Containers: func() []*model.AppContainer {
			containers := make([]*model.AppContainer, 0)
			for _, container := range entity.Containers {
				c := &model.AppContainer{
					Name:        container.Name,
					Image:       container.Image,
					PullSecret:  container.ImagePullSecret,
					ComputePlan: container.ComputePlan,
					Quantity:    container.Quantity,
					AttachedResources: func() []*model.AttachedRes {
						attached := make([]*model.AttachedRes, 0)
						for _, attachedResource := range container.AttachedResources {
							attached = append(
								attached, &model.AttachedRes{
									ResID: attachedResource.ResourceId,
								},
							)
						}
						return attached
					}(),
					EnvVars: func() []*model.EnvVar {
						envVars := make([]*model.EnvVar, 0)
						for _, envVar := range container.EnvVars {
							envVars = append(
								envVars, &model.EnvVar{
									Key: envVar.Key,
									Value: &model.EnvVal{
										Type:  envVar.Type,
										Value: envVar.Value,
									},
								},
							)
						}
						return envVars
					}(),
				}
				containers = append(containers, c)
			}
			return containers
		}(),
	}, nil
}

func (r *mutationResolver) CoreDeleteApp(ctx context.Context, appID repos.ID) (bool, error) {
	_, err := r.Domain.DeleteApp(ctx, appID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreRollbackApp(ctx context.Context, appID repos.ID, version int) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateSecret(ctx context.Context, projectID repos.ID, name string, description *string, data []*model.CSEntryIn) (*model.Secret, error) {
	entries := make([]*entities.Entry, 0)
	for _, i := range data {
		entries = append(
			entries, &entities.Entry{
				Key:   i.Key,
				Value: i.Value,
			},
		)
	}
	configEntity, err := r.Domain.CreateSecret(ctx, projectID, name, description, entries)
	if err != nil {
		return nil, err
	}
	return secretModelFromEntity(configEntity), nil
}

func (r *mutationResolver) CoreUpdateSecret(ctx context.Context, secretID repos.ID, description *string, data []*model.CSEntryIn) (bool, error) {
	entries := make([]*entities.Entry, 0)
	for _, i := range data {
		entries = append(
			entries, &entities.Entry{
				Key:   i.Key,
				Value: i.Value,
			},
		)
	}
	return r.Domain.UpdateSecret(ctx, secretID, description, entries)
}

func (r *mutationResolver) CoreDeleteSecret(ctx context.Context, secretID repos.ID) (bool, error) {
	secret, err := r.Domain.DeleteSecret(ctx, secretID)
	if err != nil {
		return false, err
	}
	return secret, nil
}

func (r *mutationResolver) CoreCreateConfig(ctx context.Context, projectID repos.ID, name string, description *string, data []*model.CSEntryIn) (*model.Config, error) {
	entries := make([]*entities.Entry, 0)
	for _, i := range data {
		entries = append(
			entries, &entities.Entry{
				Key:   i.Key,
				Value: i.Value,
			},
		)
	}
	configEntity, err := r.Domain.CreateConfig(ctx, projectID, name, description, entries)
	if err != nil {
		return nil, err
	}
	return configModelFromEntity(configEntity), nil
}

func (r *mutationResolver) CoreUpdateConfig(ctx context.Context, configID repos.ID, description *string, data []*model.CSEntryIn) (bool, error) {
	entries := make([]*entities.Entry, 0)
	for _, i := range data {
		entries = append(
			entries, &entities.Entry{
				Key:   i.Key,
				Value: i.Value,
			},
		)
	}
	return r.Domain.UpdateConfig(ctx, configID, description, entries)
}

func (r *mutationResolver) CoreDeleteConfig(ctx context.Context, configID repos.ID) (bool, error) {
	config, err := r.Domain.DeleteConfig(ctx, configID)
	if err != nil {
		return false, err
	}
	return config, nil
}

func (r *mutationResolver) CoreCreateRouter(ctx context.Context, projectID repos.ID, name string, domains []string, routes []*model.RouteInput) (*model.Router, error) {
	routeEnt := make([]*entities.Route, 0)
	for _, r := range routes {
		routeEnt = append(
			routeEnt, &entities.Route{
				Path:    r.Path,
				AppName: r.AppName,
				Port: func() uint16 {
					if r.Port != nil {
						return uint16(*r.Port)
					}
					return 0
				}(),
			},
		)
	}
	d := domains
	if domains == nil {
		d = []string{}
	}
	router, err := r.Domain.CreateRouter(ctx, projectID, name, d, routeEnt)
	if err != nil {
		return nil, err
	}
	return routerModelFromEntity(router), err
}

func (r *mutationResolver) CoreUpdateRouter(ctx context.Context, routerID repos.ID, domains []string, routes []*model.RouteInput) (bool, error) {
	entries := make([]*entities.Route, 0)
	for _, i := range routes {
		entries = append(
			entries, &entities.Route{
				Path:    i.Path,
				AppName: i.AppName,
				Port: func() uint16 {
					if i.Port != nil {
						return uint16(*i.Port)
					}
					return 0
				}(),
			},
		)
	}
	return r.Domain.UpdateRouter(ctx, routerID, domains, entries)
}

func (r *mutationResolver) CoreDeleteRouter(ctx context.Context, routerID repos.ID) (bool, error) {
	router, err := r.Domain.DeleteRouter(ctx, routerID)
	if err != nil {
		return false, err
	}
	return router, nil
}

func (r *mutationResolver) CoreCreateEdgeRegion(ctx context.Context, edgeRegion model.EdgeRegionIn, providerID repos.ID) (bool, error) {
	err := r.Domain.CreateEdgeRegion(
		ctx, providerID, &entities.EdgeRegion{
			Name:       edgeRegion.Name,
			ProviderId: providerID,
			Region:     edgeRegion.Region,
			Pools: func() []entities.NodePool {
				pools := make([]entities.NodePool, 0)
				for _, p := range edgeRegion.Pools {
					pools = append(
						pools, entities.NodePool{
							Name:   p.Name,
							Config: p.Config,
							Min:    p.Min,
							Max:    p.Max,
						},
					)
				}
				return pools
			}(),
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreUpdateEdgeRegion(ctx context.Context, edgeID repos.ID, edgeRegion model.EdgeRegionUpdateIn) (bool, error) {
	err := r.Domain.UpdateEdgeRegion(
		ctx, edgeID, &domain.EdgeRegionUpdate{
			Name: edgeRegion.Name,
			NodePools: func() []entities.NodePool {
				if edgeRegion.Pools != nil {
					pools := make([]entities.NodePool, 0)
					for _, p := range edgeRegion.Pools {
						pools = append(
							pools, entities.NodePool{
								Name:   p.Name,
								Config: p.Config,
								Min:    p.Min,
								Max:    p.Max,
							},
						)
					}
					return pools
				}
				return nil
			}(),
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreDeleteEdgeRegion(ctx context.Context, edgeID repos.ID) (bool, error) {
	err := r.Domain.DeleteEdgeRegion(ctx, edgeID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreCreateCloudProvider(ctx context.Context, accountID *repos.ID, cloudProvider model.CloudProviderIn) (bool, error) {
	err := r.Domain.CreateCloudProvider(
		ctx, accountID, &entities.CloudProvider{
			Name:      cloudProvider.Name,
			Provider:  cloudProvider.Provider,
			AccountId: accountID,
			Credentials: func() map[string]string {
				creds := make(map[string]string)
				for k, v := range cloudProvider.Credentials {
					creds[k] = v.(string)
				}
				return creds
			}(),
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreUpdateCloudProvider(ctx context.Context, providerID repos.ID, cloudProvider model.CloudProviderUpdateIn) (bool, error) {
	err := r.Domain.UpdateCloudProvider(
		ctx, providerID, &domain.CloudProviderUpdate{
			Name: cloudProvider.Name,
			Credentials: func() map[string]string {
				if cloudProvider.Credentials == nil {
					return nil
				}
				creds := make(map[string]string)
				for k, v := range cloudProvider.Credentials {
					creds[k] = v.(string)
				}
				return creds
			}(),
		},
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreDeleteCloudProvider(ctx context.Context, providerID repos.ID) (bool, error) {
	err := r.Domain.DeleteCloudProvider(ctx, providerID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreAddNewCluster(ctx context.Context, cluster model.ClusterIn) (*model.ClusterOut, error) {
	newCluster, err := r.Domain.AddNewCluster(ctx, cluster.Name, cluster.SubDomain, cluster.KubeConfig)
	if err != nil {
		return nil, err
	}
	return &model.ClusterOut{
		ID:        newCluster.Id,
		Name:      newCluster.Name,
		SubDomain: newCluster.SubDomain,
	}, nil
}

func (r *projectResolver) Memberships(ctx context.Context, obj *model.Project) ([]*model.ProjectMembership, error) {
	entities, err := r.Domain.GetProjectMemberships(ctx, obj.ID)
	accountMemberships := make([]*model.ProjectMembership, len(entities))
	for i, entity := range entities {
		accountMemberships[i] = &model.ProjectMembership{
			Project: &model.Project{
				ID: entity.ProjectId,
			},
			User: &model.User{
				ID: entity.UserId,
			},
			Role: string(entity.Role),
		}
	}
	return accountMemberships, err
}

func (r *projectResolver) DockerCredentials(ctx context.Context, obj *model.Project) (*model.DockerCredentials, error) {
	username, password, err := r.Domain.GetDockerCredentials(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	return &model.DockerCredentials{
		Username: username,
		Password: password,
	}, nil
}

func (r *projectResolver) Region(ctx context.Context, obj *model.Project) (*model.EdgeRegion, error) {
	reg, err := r.Domain.GetEdgeRegion(ctx, obj.RegionID)
	if err != nil {
		return nil, err
	}
	return &model.EdgeRegion{
		ID:        reg.Id,
		Name:      reg.Name,
		Region:    reg.Region,
		CreatedAt: reg.CreationTime.String(),
		UpdatedAt: func() *string {
			if !reg.UpdateTime.IsZero() {
				s := reg.UpdateTime.String()
				return &s
			}
			return nil
		}(),
		Pools: func() []*model.NodePool {
			pools := make([]*model.NodePool, 0)
			for _, pool := range reg.Pools {
				pools = append(
					pools, &model.NodePool{
						Name:   pool.Name,
						Config: pool.Config,
						Min:    pool.Min,
						Max:    pool.Max,
					},
				)
			}
			return pools
		}(),
	}, nil
}

func (r *queryResolver) CoreCheckDeviceExist(ctx context.Context, accountID repos.ID, name string) (bool, error) {
	exists, err := r.Domain.DeviceByNameExists(ctx, accountID, name)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *queryResolver) CoreProjects(ctx context.Context, accountID *repos.ID) ([]*model.Project, error) {
	projectEntities, err := r.Domain.GetAccountProjects(ctx, *accountID)
	if err != nil {
		return nil, err
	}
	projects := make([]*model.Project, 0)
	for _, projectEntity := range projectEntities {
		projects = append(projects, projectModelFromEntity(projectEntity))
	}
	return projects, nil
}

func (r *queryResolver) CoreProject(ctx context.Context, projectID repos.ID) (*model.Project, error) {
	p, err := r.Domain.GetProjectWithID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return projectModelFromEntity(p), nil
}

func (r *queryResolver) CoreApps(ctx context.Context, projectID repos.ID, search *string) ([]*model.App, error) {
	appEntities, err := r.Domain.GetApps(ctx, projectID)
	if err != nil {
		return nil, err
	}
	apps := returnApps(appEntities)
	return apps, nil
}

func (r *queryResolver) CoreApp(ctx context.Context, appID repos.ID) (*model.App, error) {
	a, err := r.Domain.GetApp(ctx, appID)
	if err != nil {
		return nil, err
	}
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

	return &model.App{
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
		Replicas:    &a.Replicas,
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
	}, nil
}

func (r *queryResolver) CoreGenerateEnv(ctx context.Context, projectID repos.ID, klConfig map[string]interface{}) (*model.LoadEnv, error) {
	marshal, err := json.Marshal(klConfig)
	if err != nil {
		return nil, err
	}
	var klFile localenv.KLFile
	json.Unmarshal(marshal, &klFile)

	env, m, err := r.Domain.GenerateEnv(ctx, klFile)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("%v", env)

	return &model.LoadEnv{
		EnvVars: func() map[string]any {
			mfs := make(map[string]any)
			for k, v := range env {
				mfs[k] = v
			}
			return mfs
		}(),
		MountFiles: func() map[string]any {
			mfs := make(map[string]any)
			for k, v := range m {
				mfs[k] = v
			}
			return mfs
		}(),
	}, nil
}

func (r *queryResolver) CoreRouters(ctx context.Context, projectID repos.ID, search *string) ([]*model.Router, error) {
	routerEntities, err := r.Domain.GetRouters(ctx, projectID)
	if err != nil {
		return nil, err
	}
	routers := make([]*model.Router, 0)
	for _, i := range routerEntities {
		routers = append(routers, routerModelFromEntity(i))
	}
	return routers, nil
}

func (r *queryResolver) CoreRouter(ctx context.Context, routerID repos.ID) (*model.Router, error) {
	routerEntity, err := r.Domain.GetRouter(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return routerModelFromEntity(routerEntity), nil
}

func (r *queryResolver) CoreConfigs(ctx context.Context, projectID repos.ID, search *string) ([]*model.Config, error) {
	configEntities, err := r.Domain.GetConfigs(ctx, projectID)
	if err != nil {
		return nil, err
	}
	configs := make([]*model.Config, 0)
	for _, i := range configEntities {
		configs = append(configs, configModelFromEntity(i))
	}
	return configs, nil
}

func (r *queryResolver) CoreConfig(ctx context.Context, configID repos.ID) (*model.Config, error) {
	configEntity, err := r.Domain.GetConfig(ctx, configID)
	if err != nil {
		return nil, err
	}
	return configModelFromEntity(configEntity), nil
}

func (r *queryResolver) CoreSecrets(ctx context.Context, projectID repos.ID, search *string) ([]*model.Secret, error) {
	configEntities, err := r.Domain.GetSecrets(ctx, projectID)
	if err != nil {
		return nil, err
	}
	secrets := make([]*model.Secret, 0)
	for _, i := range configEntities {
		secrets = append(secrets, secretModelFromEntity(i))
	}
	return secrets, nil
}

func (r *queryResolver) CoreSecret(ctx context.Context, secretID repos.ID) (*model.Secret, error) {
	secretEntity, err := r.Domain.GetSecret(ctx, secretID)
	if err != nil {
		return nil, err
	}
	return secretModelFromEntity(secretEntity), nil
}

func (r *queryResolver) ManagedSvcMarketList(ctx context.Context) (map[string]interface{}, error) {
	templates, err := r.Domain.GetManagedServiceTemplates(ctx)
	if err != nil {
		return nil, err
	}
	var res []any
	marshal, err := json.Marshal(templates)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(marshal, &res)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"categories": res}, nil
}

func (r *queryResolver) ManagedSvcListAvailable(ctx context.Context) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagedSvcGetInstallation(ctx context.Context, installationID repos.ID, nextVersion *bool) (*model.ManagedSvc, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagedSvcListInstallations(ctx context.Context, projectID repos.ID) ([]*model.ManagedSvc, error) {
	svcs, err := r.Domain.GetManagedSvcs(ctx, projectID)
	managedSvcs := make([]*model.ManagedSvc, 0)
	for _, svc := range svcs {
		managedSvcs = append(managedSvcs, managedSvcModelFromEntity(svc))
	}
	return managedSvcs, err
}

func (r *queryResolver) ManagedResGetResource(ctx context.Context, resID repos.ID, nextVersion *bool) (*model.ManagedRes, error) {
	res, err := r.Domain.GetManagedRes(ctx, resID)
	if err != nil {
		return nil, err
	}
	return managedResourceModelFromEntity(res), nil
}

func (r *queryResolver) ManagedResListResources(ctx context.Context, installationID repos.ID) ([]*model.ManagedRes, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreGetComputePlans(ctx context.Context) ([]*model.ComputePlan, error) {
	planEntities, err := r.Domain.GetComputePlans(ctx)
	if err != nil {
		return nil, err
	}
	plans := make([]*model.ComputePlan, 0)
	for _, i := range planEntities {
		plans = append(
			plans, &model.ComputePlan{
				Name:                  i.Name,
				Desc:                  i.Desc,
				SharingEnabled:        i.SharingEnabled,
				DedicatedEnabled:      i.DedicatedEnabled,
				MemoryPerVCPUCpu:      int(i.MemoryPerCPU),
				MaxSharedCPUPerPod:    int(i.MaxSharedCPUPerPod),
				MaxDedicatedCPUPerPod: int(i.MaxDedicatedCPUPerPod),
			},
		)
	}
	return plans, nil
}

func (r *queryResolver) CoreGetStoragePlans(ctx context.Context) ([]*model.StoragePlan, error) {
	plans, err := r.Domain.GetStoragePlans(ctx)
	if err != nil {
		return nil, err
	}
	storagePlans := make([]*model.StoragePlan, 0)
	for _, i := range plans {
		storagePlans = append(
			storagePlans, &model.StoragePlan{
				Name:        i.Name,
				Description: i.Desc,
			},
		)
	}
	return storagePlans, nil
}

func (r *queryResolver) CoreGetLamdaPlan(ctx context.Context) (*model.LambdaPlan, error) {
	return &model.LambdaPlan{Name: "Default"}, nil
}

func (r *queryResolver) CoreGetCloudProviders(ctx context.Context, accountID repos.ID) ([]*model.CloudProvider, error) {
	providers, err := r.Domain.GetCloudProviders(ctx, accountID)
	if err != nil {
		return nil, err
	}
	cloudProviders := make([]*model.CloudProvider, 0)
	for _, i := range providers {
		cloudProviders = append(
			cloudProviders, &model.CloudProvider{
				ID:       i.Id,
				Name:     i.Name,
				Provider: i.Provider,
				IsShared: *i.AccountId == "kl-core",
			},
		)
	}
	return cloudProviders, nil
}

func (r *queryResolver) CoreGetDevice(ctx context.Context, deviceID repos.ID) (*model.Device, error) {
	device, err := r.Domain.GetDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	return &model.Device{
		ID:      device.Id,
		Region:  device.ActiveRegion,
		User:    &model.User{ID: device.UserId},
		Account: &model.Account{ID: device.AccountId},
		Name:    device.Name,
		Ports: func() []*model.Port {
			var ports []*model.Port
			for _, port := range device.ExposedPorts {
				ports = append(
					ports, &model.Port{
						Port: int(port.Port),
						TargetPort: func() *int {
							if port.TargetPort != nil {
								i := int(*port.TargetPort)
								return &i
							}
							return nil
						}(),
					},
				)
			}
			return ports
		}(),
	}, nil
}

func (r *queryResolver) CoreGetEdgeNodes(ctx context.Context, edgeID repos.ID) ([]*model.EdgeNode, error) {
	nodes, err := r.Domain.GetEdgeNodes(ctx, edgeID)
	if err != nil {
		return nil, err
	}
	edgeNodes := make([]*model.EdgeNode, 0)
	for _, node := range nodes.Items {
		edgeNodes = append(
			edgeNodes, &model.EdgeNode{
				NodeIndex: node.Spec.Index,
				Status: func() (result map[string]any) {
					defer func() {
						if r := recover(); r != nil {
							result = map[string]any{}
						}
					}()
					status := make(map[string]any)
					marshal, _ := json.Marshal(node.Status)
					_ = json.Unmarshal(marshal, &status)
					return status
				}(),
				Name:         node.Name,
				Config:       node.Spec.Config,
				CreationTime: node.ObjectMeta.CreationTimestamp.Time.String(),
			},
		)
	}
	return edgeNodes, nil
}

func (r *userResolver) Devices(ctx context.Context, obj *model.User) ([]*model.Device, error) {
	var e error
	defer wErrors.HandleErr(&e)
	user := obj
	deviceEntities, e := r.Domain.ListUserDevices(ctx, user.ID)
	wErrors.AssertNoError(e, fmt.Errorf("not able to list devices of user %s", user.ID))
	devices := make([]*model.Device, 0)
	for _, device := range deviceEntities {
		devices = append(
			devices, &model.Device{
				ID:      device.Id,
				Region:  device.ActiveRegion,
				User:    &model.User{ID: device.UserId},
				Account: &model.Account{ID: device.AccountId},
				Name:    device.Name,
				Ports: func() []*model.Port {
					var ports []*model.Port
					for _, port := range device.ExposedPorts {
						ports = append(
							ports, &model.Port{
								Port: int(port.Port),
								TargetPort: func() *int {
									if port.TargetPort != nil {
										i := int(*port.TargetPort)
										return &i
									}
									return nil
								}(),
							},
						)
					}
					return ports
				}(),
			},
		)
	}
	return devices, e
}

// Account returns generated.AccountResolver implementation.
func (r *Resolver) Account() generated.AccountResolver { return &accountResolver{r} }

// App returns generated.AppResolver implementation.
func (r *Resolver) App() generated.AppResolver { return &appResolver{r} }

// CloudProvider returns generated.CloudProviderResolver implementation.
func (r *Resolver) CloudProvider() generated.CloudProviderResolver { return &cloudProviderResolver{r} }

// Device returns generated.DeviceResolver implementation.
func (r *Resolver) Device() generated.DeviceResolver { return &deviceResolver{r} }

// EdgeRegion returns generated.EdgeRegionResolver implementation.
func (r *Resolver) EdgeRegion() generated.EdgeRegionResolver { return &edgeRegionResolver{r} }

// ManagedRes returns generated.ManagedResResolver implementation.
func (r *Resolver) ManagedRes() generated.ManagedResResolver { return &managedResResolver{r} }

// ManagedSvc returns generated.ManagedSvcResolver implementation.
func (r *Resolver) ManagedSvc() generated.ManagedSvcResolver { return &managedSvcResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Project returns generated.ProjectResolver implementation.
func (r *Resolver) Project() generated.ProjectResolver { return &projectResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type accountResolver struct{ *Resolver }
type appResolver struct{ *Resolver }
type cloudProviderResolver struct{ *Resolver }
type deviceResolver struct{ *Resolver }
type edgeRegionResolver struct{ *Resolver }
type managedResResolver struct{ *Resolver }
type managedSvcResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type projectResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
