package domain

import (
	"context"
	"fmt"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetApp(ctx context.Context, appId repos.ID) (*entities.App, error) {
	return d.appRepo.FindById(ctx, appId)
}
func (d *domain) GetApps(ctx context.Context, projectID repos.ID) ([]*entities.App, error) {
	apps, err := d.appRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"project_id": projectID,
	}})
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (d *domain) OnUpdateApp(ctx context.Context, response *op_crds.StatusUpdate) error {
	one, err := d.appRepo.FindById(ctx, repos.ID(response.Metadata.ResourceId))
	if err != nil {
		return err
	}
	if response.IsReady {
		one.Status = entities.AppStateLive
	} else {
		one.Status = entities.AppStateSyncing
	}
	one.Conditions = response.ChildConditions
	_, err = d.appRepo.UpdateById(ctx, one.Id, one)
	if err != nil {
		return err
	}
	err = d.notifier.Notify(one.Id)
	if err != nil {
		return err
	}
	return err
}
func (d *domain) OnDeleteApp(ctx context.Context, name string, namespace string) error {
	one, err := d.appRepo.FindOne(ctx, repos.Filter{
		"name":      name,
		"namespace": namespace,
	})
	if err != nil {
		return err
	}
	if one == nil {
		return nil
	}
	err = d.appRepo.DeleteById(ctx, one.Id)
	if err != nil {
		return err
	}
	d.changeNotifier.Notify(one.Id)
	return nil
}

func (d *domain) InstallApp(ctx context.Context, projectId repos.ID, app entities.App) (*entities.App, error) {
	prj, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return nil, err
	}
	app.Namespace = prj.Name
	app.ProjectId = prj.Id
	app.Status = entities.AppStateSyncing
	createdApp, err := d.appRepo.Create(ctx, &app)
	if err != nil {
		return nil, err
	}

	err = d.sendAppApply(ctx, prj, createdApp)
	if err != nil {
		return nil, err
	}
	return createdApp, nil
}
func (d *domain) UpdateApp(ctx context.Context, appId repos.ID, app entities.App) (*entities.App, error) {
	prj, err := d.projectRepo.FindById(ctx, app.ProjectId)
	if err != nil {
		return nil, err
	}
	app.Namespace = prj.Name
	app.ProjectId = prj.Id
	app.Id = appId
	updatedApp, err := d.appRepo.UpdateById(ctx, appId, &app)
	if err != nil {
		return nil, err
	}
	err = d.sendAppApply(ctx, prj, updatedApp)
	if err != nil {
		return nil, err
	}
	return updatedApp, nil
}
func (d *domain) DeleteApp(ctx context.Context, appID repos.ID) (bool, error) {
	app, err := d.appRepo.FindById(ctx, appID)
	err = d.workloadMessenger.SendAction("delete", string(appID), &op_crds.App{
		APIVersion: op_crds.AppAPIVersion,
		Kind:       op_crds.AppKind,
		Metadata: op_crds.AppMetadata{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
	})
	app.Status = entities.AppStateSyncing
	_, err = d.appRepo.UpdateById(ctx, appID, app)
	d.workloadMessenger.SendAction("delete", string(appID), &op_crds.App{
		APIVersion: op_crds.AppAPIVersion,
		Kind:       op_crds.AppKind,
		Metadata: op_crds.AppMetadata{
			Name:      app.ReadableId,
			Namespace: app.Namespace,
		},
	})
	if err != nil {
		return false, err
	}
	return true, err
}

func (d *domain) sendAppApply(ctx context.Context, prj *entities.Project, app *entities.App) error {
	if app.IsLambda {
		err := d.workloadMessenger.SendAction("apply", string(app.Id), &op_crds.Lambda{
			APIVersion: op_crds.LambdaAPIVersion,
			Kind:       op_crds.LambdaKind,
			Metadata: op_crds.LambdaMetadata{
				Name:      app.ReadableId,
				Namespace: app.Namespace,
				Annotations: map[string]string{
					"kloudlite.io/account-ref":       string(prj.AccountId),
					"kloudlite.io/project-ref":       string(prj.Id),
					"kloudlite.io/resource-ref":      string(app.Id),
					"kloudlite.io/billing-plan":      "Lambda",
					"kloudlite.io/billable-quantity": fmt.Sprintf("%v", app.Containers[0].Quantity),
				},
			},
			Spec: op_crds.LambdaSpec{
				Containers: func() []op_crds.Container {
					cs := make([]op_crds.Container, 0)
					for _, c := range app.Containers {
						cs = append(cs, op_crds.Container{
							Name:  c.Name,
							Image: c.Image,
							Env: func() []op_crds.EnvEntry {
								env := make([]op_crds.EnvEntry, 0)
								for _, e := range c.EnvVars {
									if e.Type == "managed_resource" {
										ref := fmt.Sprintf("mres-%v", *e.Ref)
										env = append(env, op_crds.EnvEntry{
											Value:   e.Value,
											Key:     e.Key,
											Type:    "secret",
											RefName: &ref,
											RefKey:  e.RefKey,
										})
									} else {
										env = append(env, op_crds.EnvEntry{
											Value:   e.Value,
											Key:     e.Key,
											Type:    e.Type,
											RefName: e.Ref,
											RefKey:  e.RefKey,
										})
									}
								}
								return env
							}(),
							ResourceCpu: func() *op_crds.Limit {
								o := op_crds.Limit{
									Min: fmt.Sprintf("%vm", int(c.Quantity*500)),
									Max: fmt.Sprintf("%vm", int(c.Quantity*1000)),
								}
								return &o
							}(),
							ResourceMemory: func() *op_crds.Limit {
								plan, _ := d.getComputePlan("Basic")
								return &op_crds.Limit{
									Min: fmt.Sprintf("%vMi", int(c.Quantity*1000*(plan.MemoryPerCPU))),
									Max: fmt.Sprintf("%vMi", int(c.Quantity*1000*(plan.MemoryPerCPU))),
								}
							}(),
						})
					}
					return cs
				}(),
			},
		})
		return err
	} else {
		err := d.workloadMessenger.SendAction("apply", string(app.Id), &op_crds.App{
			APIVersion: op_crds.AppAPIVersion,
			Kind:       op_crds.AppKind,
			Metadata: op_crds.AppMetadata{
				Name:      app.ReadableId,
				Namespace: app.Namespace,
				Annotations: map[string]string{
					"kloudlite.io/account-ref":       string(prj.AccountId),
					"kloudlite.io/project-ref":       string(prj.Id),
					"kloudlite.io/resource-ref":      string(app.Id),
					"kloudlite.io/billing-plan":      app.Containers[0].ComputePlan,
					"kloudlite.io/billable-quantity": fmt.Sprintf("%v", app.Containers[0].Quantity),
					"kloudlite.io/is-shared": func() string {
						if app.Containers[0].IsShared {
							return "true"
						}
						return "false"
					}(),
				},
			},
			Spec: op_crds.AppSpec{
				Services: func() []op_crds.Service {
					svcs := make([]op_crds.Service, 0)
					for _, ep := range app.ExposedPorts {
						svcs = append(svcs, op_crds.Service{
							Port:       int(ep.Port),
							TargetPort: int(ep.TargetPort),
							Type:       string(ep.Type),
						})
					}
					return svcs
				}(),
				Hpa: func() *op_crds.HPA {
					if app.AutoScale == nil {
						return nil
					}
					return &op_crds.HPA{
						MinReplicas: int(app.AutoScale.MinReplicas),
						MaxReplicas: int(app.AutoScale.MaxReplicas),
						ThresholdCpu: func() int {
							c := app.Containers[0]
							return int(c.Quantity * float64(app.AutoScale.UsagePercentage) / float64(100) * 1000.0)
						}(),
						ThresholdMemory: func() int {
							c := app.Containers[0]
							plan, err := d.GetComputePlan(ctx, c.ComputePlan)
							if err != nil {
								panic(err)
							}
							return int(c.Quantity * 1000 * plan.MemoryPerCPU * float64(app.AutoScale.UsagePercentage))
						}(),
					}
				}(),
				Containers: func() []op_crds.Container {
					cs := make([]op_crds.Container, 0)
					for _, c := range app.Containers {
						cs = append(cs, op_crds.Container{
							Name:  c.Name,
							Image: c.Image,
							Env: func() []op_crds.EnvEntry {
								env := make([]op_crds.EnvEntry, 0)
								for _, e := range c.EnvVars {
									if e.Type == "managed_resource" {
										ref := fmt.Sprintf("mres-%v", *e.Ref)
										env = append(env, op_crds.EnvEntry{
											Value:   e.Value,
											Key:     e.Key,
											Type:    "secret",
											RefName: &ref,
											RefKey:  e.RefKey,
										})
									} else {
										env = append(env, op_crds.EnvEntry{
											Value:   e.Value,
											Key:     e.Key,
											Type:    e.Type,
											RefName: e.Ref,
											RefKey:  e.RefKey,
										})
									}
								}
								return env
							}(),
							ResourceCpu: func() *op_crds.Limit {
								o := op_crds.Limit{
									Min: fmt.Sprintf("%vm", int(c.Quantity*(func() float64 {
										if c.IsShared {
											return 500
										}
										return 1000
									})())),
									Max: fmt.Sprintf("%vm", int(c.Quantity*1000)),
								}
								return &o
							}(),
							ResourceMemory: func() *op_crds.Limit {
								plan, err := d.GetComputePlan(ctx, c.ComputePlan)
								if err != nil {
									panic(err)
								}
								return &op_crds.Limit{
									Min: fmt.Sprintf("%vMi", int(c.Quantity*1000*plan.MemoryPerCPU)),
									Max: fmt.Sprintf("%vMi", int(c.Quantity*1000*plan.MemoryPerCPU)),
								}
							}(),
						})
					}
					return cs
				}(),
				Replicas: 1,
			},
		})
		return err
	}
}
