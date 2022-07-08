package domain

import (
	"context"
	"fmt"
	"go.uber.org/fx"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/auth"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/ci"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/redpanda"
	"kloudlite.io/pkg/repos"
	rcn "kloudlite.io/pkg/res-change-notifier"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"
)

type ResourceStatus string

const (
	ResourceStatusLive       = ResourceStatus("live")
	ResourceStatusInProgress = ResourceStatus("sync-in-progress")
	ResourceStatusError      = ResourceStatus("error")
)

type domain struct {
	deviceRepo           repos.DbRepo[*entities.Device]
	clusterRepo          repos.DbRepo[*entities.Cluster]
	projectRepo          repos.DbRepo[*entities.Project]
	configRepo           repos.DbRepo[*entities.Config]
	routerRepo           repos.DbRepo[*entities.Router]
	secretRepo           repos.DbRepo[*entities.Secret]
	messageProducer      redpanda.Producer
	messageTopic         string
	logger               logger.Logger
	managedSvcRepo       repos.DbRepo[*entities.ManagedService]
	managedResRepo       repos.DbRepo[*entities.ManagedResource]
	appRepo              repos.DbRepo[*entities.App]
	managedTemplatesPath string
	workloadMessenger    WorkloadMessenger
	ciClient             ci.CIClient
	imageRepoUrlPrefix   string
	notifier             rcn.ResourceChangeNotifier
	iamClient            iam.IAMClient
	authClient           auth.AuthClient
	changeNotifier       rcn.ResourceChangeNotifier
	clusterAccountRepo   repos.DbRepo[*entities.ClusterAccount]
	financeClient        finance.FinanceClient
	computePlansPath     string
}

func (d *domain) RemoveProjectMember(ctx context.Context, projectId repos.ID, userId repos.ID) error {
	_, err := d.iamClient.RemoveMembership(ctx, &iam.InRemoveMembership{
		UserId:     string(userId),
		ResourceId: string(projectId),
	})
	if err != nil {
		return err
	}
	return nil
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

func (d *domain) getComputePlan(name string) (*entities.ComputePlan, error) {
	fileData, err := ioutil.ReadFile(d.computePlansPath)
	if err != nil {
		return nil, err
	}
	var plans []entities.ComputePlan
	err = yaml.Unmarshal(fileData, &plans)
	if err != nil {
		return nil, err
	}
	for _, plan := range plans {
		if plan.Name == name {
			return &plan, nil
		}
	}
	return nil, errors.New("plan not found")
}

func (d *domain) GetComputePlan(_ context.Context, name string) (*entities.ComputePlan, error) {
	return d.getComputePlan(name)
}

func (d *domain) GetComputePlans(_ context.Context) ([]entities.ComputePlan, error) {
	fileData, err := ioutil.ReadFile(d.computePlansPath)
	if err != nil {
		return nil, err
	}
	var plans []entities.ComputePlan
	err = yaml.Unmarshal(fileData, &plans)
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (d *domain) OnSetupClusterAccount(ctx context.Context, payload entities.SetupClusterAccountResponse) error {
	one, err := d.clusterAccountRepo.FindOne(ctx, repos.Filter{
		"cluster_id": payload.ClusterId,
		"account_id": payload.AccountId,
	})
	if err != nil {
		return err
	}
	if one == nil {
		return errors.New("cluster account not found")
	}
	if payload.Done {
		one.Status = entities.ClusterAccountStateLive
		one.WgPubKey = payload.WgPublicKey
		one.WgPort = payload.WgPort
	} else {
		one.Status = entities.ClusterAccountStateError
	}
	_, err = d.clusterAccountRepo.UpdateById(ctx, one.Id, one)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) CreateClusterAccount(ctx context.Context, data *entities.ClusterAccount, region string, provider string) (*entities.ClusterAccount, error) {
	cluster, err := d.clusterRepo.FindOne(ctx, repos.Filter{
		"region":   region,
		"provider": provider,
	})
	if err != nil {
		return nil, errors.New("No clusters available in the region")
	}
	data.ClusterID = cluster.Id
	data.Status = entities.ClusterAccountStateSyncing
	fmt.Println(repos.Filter{
		"cluster_id": cluster.Id,
		"account_id": data.AccountID,
	})
	existingAccounts, err := d.clusterAccountRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"cluster_id": cluster.Id,
		},
		Sort: map[string]interface{}{
			"index": 1,
		},
	})
	if err != nil {
		return nil, err
	}

	fmt.Println(existingAccounts, err)

	index := -1
	count := 1
	for i, d := range existingAccounts {
		fmt.Println(i, d, count)
		count++
		if d.Index != i+1 {
			index = i + 1
			break
		}
	}
	if index == -1 {
		index = count
	}

	data.Index = index
	data.WgIp = fmt.Sprintf("10.12.%d.1", index)

	create, err := d.clusterAccountRepo.Create(ctx, data)
	// TODO
	//if err != nil {
	//	return nil, err
	//}
	//err = SendAction(
	//	d.infraMessenger,
	//	entities.SetupClusterAccountAction{
	//		ClusterID: create.ClusterID,
	//		Region:    cluster.Region,
	//		Provider:  cluster.Provider,
	//		AccountId: string(create.AccountID),
	//		AccountIp: create.WgIp,
	//	},
	//)

	fmt.Println(entities.SetupClusterAccountAction{
		ClusterID: create.ClusterID,
		Region:    cluster.Region,
		Provider:  cluster.Provider,
		AccountId: string(create.AccountID),
		AccountIp: create.WgIp,
	})

	return create, nil
}

func (d *domain) UpdateResourceStatus(ctx context.Context, resourceType string, resourceNamespace string, resourceName string, status ResourceStatus) (bool, error) {
	switch resourceType {
	case "ManagedResource":
		one, err := d.managedResRepo.FindOne(ctx, repos.Filter{
			"name":      resourceName,
			"namespace": resourceNamespace,
		})
		if err != nil || one == nil {
			return false, err
		}
		if one.Status != entities.ManagedResourceStatus(status) {
			one.Status = entities.ManagedResourceStatus(status)
			_, err = d.managedResRepo.UpdateById(ctx, one.Id, one)
			if err != nil || one == nil {
				return false, err
			}
			d.changeNotifier.Notify(one.Id)
			return true, nil
		}
		return true, nil
	case "ManagedService":
		one, err := d.managedSvcRepo.FindOne(ctx, repos.Filter{
			"name":      resourceName,
			"namespace": resourceNamespace,
		})
		if err != nil || one == nil {
			return false, err
		}
		if one.Status != entities.ManagedServiceStatus(status) {
			one.Status = entities.ManagedServiceStatus(status)
			_, err = d.managedSvcRepo.UpdateById(ctx, one.Id, one)
			if err != nil {
				return false, err
			}
			d.changeNotifier.Notify(one.Id)
			return true, nil
		}
		return true, nil
	case "App":
		one, err := d.appRepo.FindOne(ctx, repos.Filter{
			"readable_id": resourceName,
			"namespace":   resourceNamespace,
		})
		if err != nil || one == nil {
			fmt.Println(err)
			return false, err
		}
		if one.Status != entities.AppStatus(status) {
			one.Status = entities.AppStatus(status)
			_, err = d.appRepo.UpdateById(ctx, one.Id, one)
			if err != nil {
				return false, err
			}
			d.changeNotifier.Notify(one.Id)
			return true, nil
		}
		return true, nil
	case "Router":
		one, err := d.routerRepo.FindOne(ctx, repos.Filter{
			"name":      resourceName,
			"namespace": resourceNamespace,
		})
		if err != nil || one == nil {
			return false, err
		}
		if one.Status != entities.RouterStatus(status) {
			one.Status = entities.RouterStatus(status)
			_, err = d.routerRepo.UpdateById(ctx, one.Id, one)
			if err != nil {
				return false, err
			}
			d.changeNotifier.Notify(one.Id)
			return true, nil
		}
		return true, nil
	default:
		return false, errors.New("unsupported resource type")
	}
}

func (d *domain) GetProjectMemberships(ctx context.Context, projectID repos.ID) ([]*entities.ProjectMembership, error) {
	rbs, err := d.iamClient.ListResourceMemberships(ctx, &iam.InResourceMemberships{
		ResourceId:   string(projectID),
		ResourceType: string(common.ResourceProject),
	})
	if err != nil {
		return nil, err
	}
	var memberships []*entities.ProjectMembership
	for _, rb := range rbs.RoleBindings {
		memberships = append(memberships, &entities.ProjectMembership{
			ProjectId: repos.ID(rb.ResourceId),
			UserId:    repos.ID(rb.UserId),
			Role:      common.Role(rb.Role),
		})
	}

	if err != nil {
		return nil, err
	}
	return memberships, nil
}

func (d *domain) InviteProjectMember(ctx context.Context, projectID repos.ID, email string, role string) (bool, error) {
	byEmail, err := d.authClient.EnsureUserByEmail(ctx, &auth.GetUserByEmailRequest{Email: email})
	if err != nil {
		return false, err
	}
	if byEmail == nil {
		return false, errors.New("user not found")
	}
	_, err = d.iamClient.InviteMembership(ctx, &iam.InAddMembership{
		UserId:       byEmail.UserId,
		ResourceType: "project",
		ResourceId:   string(projectID),
		Role:         role,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) GetResourceOutputs(ctx context.Context, managedResID repos.ID) (map[string]string, error) {
	mres, err := d.managedResRepo.FindById(ctx, managedResID)
	if err != nil {
		return nil, err
	}
	project, err := d.projectRepo.FindById(ctx, mres.ProjectId)
	if err != nil {
		return nil, err
	}
	_, err = d.clusterRepo.FindOne(ctx, repos.Filter{
		"account_id": project.AccountId,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
	// TODO
	//output, err := d.infraClient.GetResourceOutput(ctx, &infra.GetInput{
	//	ManagedResName: mres.Name,
	//	ClusterId:      string(cluster.Id),
	//	Namespace:      mres.Namespace,
	//})
	//if err != nil {
	//	fmt.Println(err)
	//	return nil, err
	//}
	//return output.Output, err
}

// func (d *domain) createApp(ctx context.Context, app entities.App) (*entities.App, error) {
// 	a := entities.App{
// 		ReadableId:   app.ReadableId,
// 		ProjectId:    app.ProjectId,
// 		Name:         app.Name,
// 		Namespace:    app.Namespace,
// 		Description:  app.Description,
// 		Replicas:     app.Replicas,
// 		ExposedPorts: app.ExposedPorts,
// 		Status:       entities.AppStateSyncing,
// 	}
// 	for _, c := range app.Containers {
// 		var iName string
// 		if c.Image != nil {
// 			iName = fmt.Sprintf("%s:latest", *c.Image)
// 		}
// 		plan, err := d.getComputePlan(c.ComputePlan)

// 		if err != nil {
// 			return nil, err
// 		}
// 		container := entities.Container{
// 			Name:            c.Name,
// 			Image:           &iName,
// 			ImagePullSecret: c.ImagePullSecret,
// 			EnvVars:         c.EnvVars,
// 			CPULimits: entities.Limit{
// 				Min: fmt.Sprintf("%v%v", c.ComputePlanQuantity*1000, "m"),
// 				Max: fmt.Sprintf("%v%v", c.ComputePlanQuantity*1000, "m"),
// 			},
// 			MemoryLimits: entities.Limit{
// 				Min: fmt.Sprintf("%v%v", plan.MemoryPerCPU*c.ComputePlanQuantity, "Gi"),
// 				Max: fmt.Sprintf("%v%v", plan.MemoryPerCPU*c.ComputePlanQuantity, "Gi"),
// 			},
// 			AttachedResources: c.AttachedResources,
// 		}
// 		if c.SharingEnabled {
// 			container.CPULimits.Min = fmt.Sprintf("%v%v", c.ComputePlanQuantity*1000/2, "m")
// 		}

// 		a.Containers = append(a.Containers, container)
// 	}
// 	return &a, nil
// }

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

	// svcs := make([]op_crds.Service, 0)
	// for _, ep := range app.ExposedPorts {
	// 	svcs = append(svcs, op_crds.Service{
	// 		Port:       int(ep.Port),
	// 		TargetPort: int(ep.TargetPort),
	// 		Type:       string(ep.Type),
	// 	})
	// }
	// containers := make([]op_crds.Container, 0)
	// for _, c := range app.Containers {
	// 	env := make([]op_crds.EnvEntry, 0)
	// 	for _, e := range c.EnvVars {
	// 		if e.Type == "managed_resource" {
	// 			ref := fmt.Sprintf("mres-%v", *e.Ref)
	// 			env = append(env, op_crds.EnvEntry{
	// 				Value:   e.Value,
	// 				Key:     e.Key,
	// 				Type:    "secret",
	// 				RefName: &ref,
	// 				RefKey:  e.RefKey,
	// 			})
	// 		} else {
	// 			env = append(env, op_crds.EnvEntry{
	// 				Value:   e.Value,
	// 				Key:     e.Key,
	// 				Type:    e.Type,
	// 				RefName: e.Ref,
	// 				RefKey:  e.RefKey,
	// 			})
	// 		}
	// 	}
	// 	containers = append(containers, op_crds.Container{
	// 		Name:  c.Name,
	// 		Image: c.Image,
	// 		Env:   env,
	// 	})
	// }

	// d.workloadMessenger.SendAction("apply", string(app.Id), &op_crds.App{
	// 	APIVersion: op_crds.AppAPIVersion,
	// 	Kind:       op_crds.AppKind,
	// 	Metadata: op_crds.AppMetadata{
	// 		Name:      app.ReadableId,
	// 		Namespace: app.Namespace,
	// 	},
	// 	Spec: op_crds.AppSpec{
	// 		Services:   svcs,
	// 		Containers: containers,
	// 		Replicas:   1,
	// 	},
	// })

	return updatedApp, nil
}

func (d *domain) InstallApp(
	ctx context.Context,
	projectId repos.ID,
	app entities.App,
) (*entities.App, error) {
	prj, err := d.projectRepo.FindById(ctx, projectId)

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	app.Namespace = prj.Name
	app.ProjectId = prj.Id
	createdApp, err := d.appRepo.Create(ctx, &app)
	if err != nil {
		return nil, err
	}

	if app.IsLambda {
		d.workloadMessenger.SendAction("apply", string(app.Id), &op_crds.Lambda{
			APIVersion: op_crds.LambdaAPIVersion,
			Kind:       op_crds.LambdaKind,
			Metadata: op_crds.LambdaMetadata{
				Name:      app.ReadableId,
				Namespace: app.Namespace,
				Annotations: map[string]string{
					"kloudlite.io/account-ref":       string(prj.AccountId),
					"kloudlite.io/project-ref":       string(projectId),
					"kloudlite.io/resource-ref":      string(createdApp.Id),
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
									Min: int(c.Quantity * 500),
									Max: int(c.Quantity * 1000),
								}
								return &o
							}(),
							ResourceMemory: func() *op_crds.Limit {
								plan, _ := d.getComputePlan("Basic")
								return &op_crds.Limit{
									Min: int(c.Quantity * 1000 * (plan.MemoryPerCPU)),
									Max: int(c.Quantity * 1000 * (plan.MemoryPerCPU)),
								}
							}(),
						})
					}
					return cs
				}(),
			},
		})
	} else {
		d.workloadMessenger.SendAction("apply", string(app.Id), &op_crds.App{
			APIVersion: op_crds.AppAPIVersion,
			Kind:       op_crds.AppKind,
			Metadata: op_crds.AppMetadata{
				Name:      app.ReadableId,
				Namespace: app.Namespace,
				Annotations: map[string]string{
					"kloudlite.io/account-ref":       string(prj.AccountId),
					"kloudlite.io/project-ref":       string(projectId),
					"kloudlite.io/resource-ref":      string(createdApp.Id),
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
									Min: int(c.Quantity * (func() float64 {
										if c.IsShared {
											return 500
										}
										return 1000
									})()),
									Max: int(c.Quantity * 1000),
								}
								fmt.Printf("o: %+v\n", o)
								return &o
							}(),
							ResourceMemory: func() *op_crds.Limit {
								fmt.Println("plan:", c.ComputePlan)
								plan, err := d.GetComputePlan(ctx, c.ComputePlan)
								if err != nil {
									panic(err)
								}
								return &op_crds.Limit{
									Min: int(c.Quantity * 1000 * plan.MemoryPerCPU),
									Max: int(c.Quantity * 1000 * plan.MemoryPerCPU),
								}
							}(),
						})
					}
					return cs
				}(),
				Replicas: 1,
			},
		})
	}

	return createdApp, nil
}

func (d *domain) GetDeviceConfig(ctx context.Context, deviceId repos.ID) (string, error) {
	device, err := d.deviceRepo.FindById(ctx, deviceId)
	if err != nil {
		return "", err
	}
	cluster, err := d.clusterRepo.FindById(ctx, device.ClusterId)
	clusterAccount, err := d.clusterAccountRepo.FindOne(ctx, repos.Filter{
		"cluster_id": device.ClusterId,
		"account_id": device.AccountId,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
[Interface]
PrivateKey = %v
Address = %v/32
DNS = 10.43.0.10

[Peer]
PublicKey = %v
AllowedIPs = 10.42.0.0/16, 10.43.0.0/16, 10.13.13.0/24
Endpoint = %v:%v
`, *device.PrivateKey, device.Ip, clusterAccount.WgPubKey, cluster.Name, clusterAccount.WgPort), nil
}

func (d *domain) GetManagedServiceTemplates(ctx context.Context) ([]*entities.ManagedServiceCategory, error) {
	templates := make([]*entities.ManagedServiceCategory, 0)
	data, err := os.ReadFile(d.managedTemplatesPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &templates)
	if err != nil {
		return nil, err
	}
	fmt.Println(templates)
	return templates, nil
}

func isReady(c []metav1.Condition) bool {
	for _, _c := range c {
		if _c.Type == "Ready" && _c.Status == "True" {
			return true
		}
	}
	return false
}

func (d *domain) OnUpdateProject(ctx context.Context, response *op_crds.Project) error {
	one, err := d.projectRepo.FindOne(ctx, repos.Filter{
		"name": response.Metadata.Name,
		//"cluster_id": response.ClusterId,
	})
	if err != nil {
		return err
	}
	if isReady(response.Status.Conditions) {
		one.Status = entities.ProjectStateLive
	} else {
		one.Status = entities.ProjectStateSyncing
	}
	one.Conditions = response.Status.Conditions
	_, err = d.projectRepo.UpdateById(ctx, one.Id, one)
	return err
}

func (d *domain) OnUpdateConfig(ctx context.Context, configId repos.ID) error {
	one, err := d.configRepo.FindById(ctx, configId)
	if err != nil {
		return err
	}
	one.Status = entities.ConfigStateLive
	_, err = d.configRepo.UpdateById(ctx, one.Id, one)

	return err
}

func (d *domain) OnUpdateSecret(ctx context.Context, secretId repos.ID) error {
	one, err := d.secretRepo.FindById(ctx, secretId)
	if err != nil {
		return err
	}
	one.Status = entities.SecretStateLive
	_, err = d.secretRepo.UpdateById(ctx, one.Id, one)
	return err
}

func (d *domain) OnUpdateRouter(ctx context.Context, response *op_crds.Router) error {
	one, err := d.routerRepo.FindOne(ctx, repos.Filter{
		"name":      response.Metadata.Name,
		"namespace": response.Metadata.Namespace,
	})
	if err != nil {
		return err
	}
	if isReady(response.Status.Conditions) {
		one.Status = entities.RouteStateLive
	} else {
		one.Status = entities.RouteStateSyncing
	}
	one.Conditions = response.Status.Conditions
	_, err = d.routerRepo.UpdateById(ctx, one.Id, one)
	return err
}

func (d *domain) OnUpdateManagedSvc(ctx context.Context, response *op_crds.ManagedService) error {
	one, err := d.managedSvcRepo.FindOne(ctx, repos.Filter{
		"name":      response.Metadata.Name,
		"namespace": response.Metadata.Namespace,
	})
	if err != nil {
		return err
	}
	if isReady(response.Status.Conditions) {
		one.Status = entities.ManagedServiceStateLive
	} else {
		one.Status = entities.ManagedServiceStateSyncing
	}
	one.Conditions = response.Status.Conditions
	_, err = d.managedSvcRepo.UpdateById(ctx, one.Id, one)
	return err
}

func (d *domain) OnUpdateManagedRes(ctx context.Context, response *op_crds.ManagedResource) error {
	one, err := d.managedResRepo.FindOne(ctx, repos.Filter{
		"name":      response.Metadata.Name,
		"namespace": response.Metadata.Namespace,
	})
	if err != nil {
		return err
	}
	if isReady(response.Status.Conditions) {
		one.Status = entities.ManagedResourceStateLive
	} else {
		one.Status = entities.ManagedResourceStateSyncing
	}
	one.Conditions = response.Status.Conditions
	_, err = d.managedResRepo.UpdateById(ctx, one.Id, one)
	return err
}

func (d *domain) OnUpdateApp(ctx context.Context, response *op_crds.App) error {
	one, err := d.appRepo.FindOne(ctx, repos.Filter{
		"name":      response.Metadata.Name,
		"namespace": response.Metadata.Namespace,
	})
	if err != nil {
		return err
	}
	if isReady(response.Status.Conditions) {
		one.Status = entities.AppStateLive
	} else {
		one.Status = entities.AppStateSyncing
	}
	one.Conditions = response.Status.Conditions
	_, err = d.appRepo.UpdateById(ctx, one.Id, one)
	return err
}

func (d *domain) PatchConfig(ctx context.Context, configId repos.ID, desc *string, configData []*entities.Entry) (bool, error) {
	one, err := d.configRepo.FindById(ctx, configId)
	if err != nil {
		return false, err
	}
	if desc != nil {
		one.Description = desc
	}
	if configData != nil {
		if one.Data == nil {
			one.Data = make([]*entities.Entry, 0)
		}
		for _, v := range configData {
			inserted := false
			for _, v2 := range make([]*entities.Entry, 0) {
				if v.Key == v2.Key {
					v2.Value = v.Value
					inserted = true
					break
				}
			}
			if !inserted {
				one.Data = append(one.Data, v)
			}
		}
	}
	_, err = d.configRepo.UpdateById(ctx, one.Id, one)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) DeleteConfig(ctx context.Context, configId repos.ID) (bool, error) {
	err := d.configRepo.DeleteById(ctx, configId)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) PatchSecret(ctx context.Context, secretId repos.ID, desc *string, secretData []*entities.Entry) (bool, error) {
	one, err := d.secretRepo.FindById(ctx, secretId)
	if err != nil {
		return false, err
	}
	if desc != nil {
		one.Description = desc
	}
	if secretData != nil {
		if one.Data == nil {
			one.Data = make([]*entities.Entry, 0)
		}
		for _, v := range secretData {
			inserted := false
			for _, v2 := range make([]*entities.Entry, 0) {
				if v.Key == v2.Key {
					v2.Value = v.Value
					inserted = true
					break
				}
			}
			if !inserted {
				one.Data = append(one.Data, v)
			}
		}
	}
	_, err = d.secretRepo.UpdateById(ctx, one.Id, one)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) DeleteSecret(ctx context.Context, secretId repos.ID) (bool, error) {
	err := d.secretRepo.DeleteById(ctx, secretId)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) DeleteRouter(ctx context.Context, routerID repos.ID) (bool, error) {
	err := d.secretRepo.DeleteById(ctx, routerID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) GetManagedSvc(ctx context.Context, managedSvcID repos.ID) (*entities.ManagedService, error) {
	return d.managedSvcRepo.FindById(ctx, managedSvcID)
}

func (d *domain) GetManagedSvcs(ctx context.Context, projectID repos.ID) ([]*entities.ManagedService, error) {
	return d.managedSvcRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"project_id": projectID,
	}})
}

func (d *domain) GetManagedRes(ctx context.Context, managedResID repos.ID) (*entities.ManagedResource, error) {
	return d.managedResRepo.FindById(ctx, managedResID)
}

func (d *domain) GetManagedResources(ctx context.Context, projectID repos.ID) ([]*entities.ManagedResource, error) {
	return d.managedResRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"project_id": projectID,
	}})
}

func (d *domain) GetManagedResourcesOfService(
	ctx context.Context,
	installationId repos.ID,
) ([]*entities.ManagedResource, error) {
	fmt.Println("GetManagedResourcesOfService", installationId)
	return d.managedResRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"service_id": installationId,
	}})
}

func (d *domain) InstallManagedRes(
	ctx context.Context,
	installationId repos.ID,
	name string,
	resourceType string,
	values map[string]string,
) (*entities.ManagedResource, error) {
	svc, err := d.managedSvcRepo.FindById(ctx, installationId)
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, fmt.Errorf("managed service not found")
	}

	prj, err := d.projectRepo.FindById(ctx, svc.ProjectId)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}

	create, err := d.managedResRepo.Create(ctx, &entities.ManagedResource{
		ProjectId:    prj.Id,
		Namespace:    prj.Name,
		ServiceId:    svc.Id,
		ResourceType: entities.ManagedResourceType(resourceType),
		Name:         name,
		Values:       values,
	})
	if err != nil {
		return nil, err
	}

	d.workloadMessenger.SendAction("apply", string(create.Id), &op_crds.ManagedResource{
		APIVersion: op_crds.ManagedResourceAPIVersion,
		Kind:       op_crds.ManagedResourceKind,
		Metadata: op_crds.ManagedResourceMetadata{
			Name:      create.Name,
			Namespace: create.Namespace,
		},
		Spec: op_crds.ManagedResourceSpec{
			ManagedService: svc.Name,
			Type:           resourceType,
			Inputs:         create.Values,
		},
		Status: op_crds.Status{},
	})

	return create, nil
}

func (d *domain) UpdateManagedRes(ctx context.Context, managedResID repos.ID, values map[string]string) (bool, error) {
	id, err := d.managedResRepo.FindById(ctx, managedResID)
	if err != nil {
		return false, err
	}
	id.Values = values
	_, err = d.managedResRepo.UpdateById(ctx, managedResID, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) UnInstallManagedRes(ctx context.Context, appID repos.ID) (bool, error) {
	err := d.managedResRepo.DeleteById(ctx, appID)
	if err != nil {
		return false, err
	}
	return true, err
}

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

func (d *domain) InstallManagedSvc(ctx context.Context, projectID repos.ID, templateID repos.ID, name string, values map[string]interface{}) (*entities.ManagedService, error) {
	prj, err := d.projectRepo.FindById(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}

	create, err := d.managedSvcRepo.Create(ctx, &entities.ManagedService{
		Name:        name,
		Namespace:   prj.Name,
		ProjectId:   prj.Id,
		ServiceType: entities.ManagedServiceType(templateID),
		Values:      values,
		Status:      entities.ManagedServiceStateSyncing,
	})
	if err != nil {
		return nil, err
	}

	vs := make(map[string]string, 0)

	for k, v := range create.Values {
		vs[k] = v.(string)
	}

	err = d.workloadMessenger.SendAction("apply", string(create.Id), &op_crds.ManagedService{
		APIVersion: op_crds.ManagedServiceAPIVersion,
		Kind:       op_crds.ManagedServiceKind,
		Metadata: op_crds.ManagedServiceMetadata{
			Name:      create.Name,
			Namespace: create.Namespace,
		},
		Spec: op_crds.ManagedServiceSpec{
			Type:   "MongoDBStandalone",
			Inputs: vs,
		},
	})
	if err != nil {
		return nil, err
	}

	return create, err
}

func (d *domain) UpdateManagedSvc(ctx context.Context, managedServiceId repos.ID, values map[string]interface{}) (bool, error) {
	managedSvc, err := d.managedSvcRepo.FindById(ctx, managedServiceId)
	if err != nil {
		return false, err
	}
	if managedSvc == nil {
		return false, fmt.Errorf("project not found")
	}
	managedSvc.Values = values
	managedSvc.Status = entities.ManagedServiceStateSyncing
	_, err = d.managedSvcRepo.UpdateById(ctx, managedServiceId, managedSvc)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) UnInstallManagedSvc(ctx context.Context, managedServiceId repos.ID) (bool, error) {
	err := d.managedSvcRepo.DeleteById(ctx, managedServiceId)
	// TODO send message
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) UpdateRouter(ctx context.Context, id repos.ID, domains []string, entries []*entities.Route) (bool, error) {
	router, err := d.routerRepo.FindById(ctx, id)
	if err != nil {
		return false, err
	}
	if domains != nil {
		router.Domains = domains
	}
	if entries != nil {
		router.Routes = entries
	}
	_, err = d.routerRepo.UpdateById(ctx, id, router)
	if err != nil {
		return false, err
	}
	rs := make([]op_crds.Route, 0)
	for _, r := range router.Routes {
		rs = append(rs, op_crds.Route{
			Path: r.Path,
			App:  r.AppName,
			Port: r.Port,
		})
	}
	err = d.workloadMessenger.SendAction("apply", string(router.Id), op_crds.Router{
		APIVersion: op_crds.RouterAPIVersion,
		Kind:       op_crds.RouterKind,
		Metadata: op_crds.RouterMetadata{
			Name:      router.Name,
			Namespace: router.Namespace,
		},
		Spec: op_crds.RouterSpec{
			Domains: router.Domains,
			Routes:  rs,
		},
		Status: op_crds.Status{},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) GetRouter(ctx context.Context, routerID repos.ID) (*entities.Router, error) {
	router, err := d.routerRepo.FindById(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return router, nil
}

func (d *domain) GetRouters(ctx context.Context, projectID repos.ID) ([]*entities.Router, error) {
	routers, err := d.routerRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"project_id": projectID,
		},
	})
	if err != nil {
		return nil, err
	}
	return routers, nil
}

func (d *domain) CreateRouter(ctx context.Context, projectId repos.ID, routerName string, domains []string, routes []*entities.Route) (*entities.Router, error) {
	prj, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}
	create, err := d.routerRepo.Create(ctx, &entities.Router{
		ProjectId: projectId,
		Name:      routerName,
		Namespace: prj.Name,
		Domains:   domains,
		Routes:    routes,
	})
	if err != nil {
		return nil, err
	}

	rs := make([]op_crds.Route, 0)
	for _, r := range routes {
		rs = append(rs, op_crds.Route{
			Path: r.Path,
			App:  r.AppName,
			Port: r.Port,
		})
	}
	return create, nil
}

func (d *domain) CreateSecret(ctx context.Context, projectId repos.ID, secretName string, desc *string, secretData []*entities.Entry) (*entities.Secret, error) {
	prj, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}
	create, err := d.secretRepo.Create(ctx, &entities.Secret{
		Name:        strings.ToLower(secretName),
		ProjectId:   projectId,
		Namespace:   prj.Name,
		Data:        secretData,
		Description: desc,
	})
	if err != nil {
		return nil, err
	}
	return create, nil
}

func (d *domain) UpdateSecret(ctx context.Context, secretId repos.ID, desc *string, secretData []*entities.Entry) (bool, error) {
	cfg, err := d.secretRepo.FindById(ctx, secretId)
	if err != nil {
		return false, err
	}
	if cfg == nil {
		return false, fmt.Errorf("config not found")
	}
	if desc != nil {
		cfg.Description = desc
	}
	cfg.Data = secretData
	_, err = d.secretRepo.UpdateById(ctx, secretId, cfg)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) GetSecrets(ctx context.Context, projectId repos.ID) ([]*entities.Secret, error) {
	secrets, err := d.secretRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"project_id": projectId,
		},
	})
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (d *domain) GetSecret(ctx context.Context, secretId repos.ID) (*entities.Secret, error) {
	sec, err := d.secretRepo.FindById(ctx, secretId)
	if err != nil {
		return nil, err
	}
	return sec, nil
}

func (d *domain) GetConfig(ctx context.Context, configId repos.ID) (*entities.Config, error) {
	cfg, err := d.configRepo.FindById(ctx, configId)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (d *domain) GetConfigs(ctx context.Context, projectId repos.ID) ([]*entities.Config, error) {
	configs, err := d.configRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"project_id": projectId,
		},
	})
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func (d *domain) UpdateConfig(ctx context.Context, configId repos.ID, desc *string, configData []*entities.Entry) (bool, error) {
	cfg, err := d.configRepo.FindById(ctx, configId)
	if err != nil {
		return false, err
	}
	if cfg == nil {
		return false, fmt.Errorf("config not found")
	}
	if desc != nil {
		cfg.Description = desc
	}
	cfg.Data = configData
	cfg.Status = entities.ConfigStateSyncing
	_, err = d.configRepo.UpdateById(ctx, configId, cfg)
	if err != nil {
		return false, err
	}
	m := make(map[string]any, 0)
	for _, i := range cfg.Data {
		m[i.Key] = i.Value
	}
	err = d.workloadMessenger.SendAction("apply", string(cfg.Id), op_crds.Config{
		APIVersion: op_crds.ConfigAPIVersion,
		Kind:       op_crds.ConfigKind,
		Metadata: op_crds.ConfigMetadata{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
		},
		Data: m,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) CreateConfig(ctx context.Context, projectId repos.ID, configName string, desc *string, configData []*entities.Entry) (*entities.Config, error) {
	prj, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}
	create, err := d.configRepo.Create(ctx, &entities.Config{
		Name:        strings.ToLower(configName),
		ProjectId:   projectId,
		Namespace:   prj.Name,
		Data:        configData,
		Description: desc,
	})
	if err != nil {
		return nil, err
	}
	err = d.workloadMessenger.SendAction("apply", string(create.Id), op_crds.Config{
		APIVersion: op_crds.ConfigAPIVersion,
		Kind:       op_crds.ConfigKind,
		Metadata: op_crds.ConfigMetadata{
			Name:      configName,
			Namespace: prj.Name,
		},
		Data: nil,
	})
	fmt.Println("scheduled")
	time.AfterFunc(3*time.Second, func() {
		fmt.Println("send apply config")
		d.notifier.Notify(create.Id)
	})
	if err != nil {
		return nil, err
	}
	return create, nil
}

func (d *domain) GetProjectWithID(ctx context.Context, projectId repos.ID) (*entities.Project, error) {
	id, err := d.projectRepo.FindById(ctx, projectId)
	return id, err
}

func (d *domain) GetAccountProjects(ctx context.Context, acountId repos.ID) ([]*entities.Project, error) {
	res, err := d.projectRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"account_id": acountId,
		},
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func generateReadable(name string) string {
	compile := regexp.MustCompile(`[^0-9a-zA-Z:,/s]+`)
	allString := compile.ReplaceAllString(strings.ToLower(name), "")
	m := math.Min(10, float64(len(allString)))
	return fmt.Sprintf("%v_%v", allString[:int(m)], rand.Intn(9999))
}

func (d *domain) CreateProject(ctx context.Context, ownerId repos.ID, accountId repos.ID, projectName string, displayName string, logo *string, cluster string, description *string) (*entities.Project, error) {
	create, err := d.projectRepo.Create(ctx, &entities.Project{
		Name:        projectName,
		AccountId:   accountId,
		ReadableId:  repos.ID(generateReadable(projectName)),
		DisplayName: displayName,
		Logo:        logo,
		Description: description,
		Cluster:     cluster,
		Status:      entities.ProjectStateSyncing,
	})
	if err != nil {
		return nil, err
	}
	_, err = d.iamClient.AddMembership(ctx, &iam.InAddMembership{
		UserId:       string(ownerId),
		ResourceType: "project",
		ResourceId:   string(create.Id),
		Role:         "owner",
	})
	if err != nil {
		return nil, err
	}
	err = d.workloadMessenger.SendAction("apply", string(create.Id), &op_crds.Project{
		APIVersion: op_crds.APIVersion,
		Kind:       op_crds.ProjectKind,
		Metadata: op_crds.ProjectMetadata{
			Name: create.Name,
			Annotations: map[string]string{
				"kloudlite.io/account-ref": string(accountId),
			},
		},
		Spec: op_crds.ProjectSpec{
			DisplayName: displayName,
			ArtifactRegistry: op_crds.ArtifactRegistry{
				Enabled: false,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return create, err
}

func (d *domain) OnSetupCluster(ctx context.Context, response entities.SetupClusterResponse) error {
	fmt.Println("OnSetupClusterDone")
	byId, err := d.clusterRepo.FindById(ctx, response.ClusterID)
	if err != nil {
		return err
	}
	if response.Done {
		byId.Status = entities.ClusterStateLive
	} else {
		byId.Status = entities.ClusterStateError
	}

	_, err = d.clusterRepo.UpdateById(ctx, response.ClusterID, byId)
	if err != nil {
		panic(err)
		return err
	}
	return nil
}

func (d *domain) OnDeleteCluster(cxt context.Context, response entities.DeleteClusterResponse) error {
	byId, err := d.clusterRepo.FindById(cxt, response.ClusterID)
	if err != nil {
		return err
	}
	if response.Done {
		byId.Status = entities.ClusterStateDown
	} else {
		byId.Status = entities.ClusterStateError
	}
	_, err = d.clusterRepo.UpdateById(cxt, response.ClusterID, byId)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) OnUpdateCluster(cxt context.Context, response entities.UpdateClusterResponse) error {
	byId, err := d.clusterRepo.FindById(cxt, response.ClusterID)
	if err != nil {
		return err
	}
	if response.Done {
		byId.Status = entities.ClusterStateLive
	} else {
		byId.Status = entities.ClusterStateError
	}
	_, err = d.clusterRepo.UpdateById(cxt, response.ClusterID, byId)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) OnAddPeer(cxt context.Context, response entities.AddPeerResponse) error {
	byId, err := d.deviceRepo.FindById(cxt, response.DeviceID)
	if err != nil {
		return err
	}
	if response.Done {
		byId.Status = entities.DeviceStateAttached
	} else {
		byId.Status = entities.DeviceStateError
	}
	_, err = d.deviceRepo.UpdateById(cxt, response.ClusterID, byId)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) OnDeletePeer(cxt context.Context, response entities.DeletePeerResponse) error {
	byId, err := d.deviceRepo.FindById(cxt, response.DeviceID)
	if err != nil {
		return err
	}
	if response.Done {
		byId.Status = entities.DeviceStateDeleted
	} else {
		byId.Status = entities.DeviceStateError
	}
	_, err = d.deviceRepo.UpdateById(cxt, response.ClusterID, byId)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) CreateCluster(ctx context.Context, data *entities.Cluster) (cluster *entities.Cluster, e error) {
	data.Status = entities.ClusterStateSyncing
	c, err := d.clusterRepo.Create(ctx, data)
	if err != nil {
		return nil, err
	}
	fmt.Println(entities.SetupClusterAction{
		ClusterID: c.Id,
		Region:    c.Region,
		Provider:  c.Provider,
	})
	// TODO
	//err = SendAction(
	//	d.infraMessenger,
	//	entities.SetupClusterAction{
	//		ClusterID:  c.Id,
	//		Region:     c.Region,
	//		Provider:   c.Provider,
	//		NodesCount: c.NodesCount,
	//	},
	//)
	if err != nil {
		panic(err)
		return nil, err
	}
	return c, e
}

func (d *domain) UpdateCluster(
	ctx context.Context,
	id repos.ID,
	name *string,
	nodeCount *int,
) (bool, error) {
	c, err := d.clusterRepo.FindById(ctx, id)
	if c.Status == entities.ClusterStateSyncing {
		return false, errors.New("Cluster is still syncing")
	}
	if err != nil {
		return false, err
	}
	if name != nil {
		c.Name = *name
	}
	if nodeCount != nil {
		c.Status = entities.ClusterStateSyncing
	}
	_, err = d.clusterRepo.UpdateById(ctx, id, c)
	if err != nil {
		return false, err
	}
	if c.Status == entities.ClusterStateSyncing {
		// TODO
		//err = SendAction(d.infraMessenger, entities.UpdateClusterAction{
		//	ClusterID:  id,
		//	Region:     updated.Region,
		//	Provider:   updated.Provider,
		//	NodesCount: updated.NodesCount,
		//})
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (d *domain) DeleteCluster(ctx context.Context, clusterId repos.ID) error {
	cluster, err := d.clusterRepo.FindById(ctx, clusterId)
	if err != nil {
		return err
	}
	cluster.Status = entities.ClusterStateSyncing
	_, err = d.clusterRepo.UpdateById(ctx, clusterId, cluster)
	if err != nil {
		return err
	}
	//err = SendAction(d.infraMessenger, entities.DeleteClusterAction{
	//	ClusterID: clusterId,
	//	Provider:  cluster.Provider,
	//})
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) ListClusterSubscriptions(ctx context.Context, accountId repos.ID) ([]*entities.ClusterAccount, error) {
	return d.clusterAccountRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"account_id": accountId,
		},
	})
}

func (d *domain) AddDevice(ctx context.Context, deviceName string, clusterId repos.ID, accountId repos.ID, userId repos.ID) (*entities.Device, error) {

	clusterAccount, err := d.clusterAccountRepo.FindOne(ctx, repos.Filter{
		"cluster_id": clusterId,
		"account_id": accountId,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch account on cluster %v", err)
	}

	pk, e := wgtypes.GeneratePrivateKey()
	if e != nil {
		return nil, fmt.Errorf("unable to generate private key because %v", e)
	}
	pkString := pk.String()
	pbKeyString := pk.PublicKey().String()

	devices, err := d.deviceRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"cluster_id": clusterId,
		},
		Sort: map[string]any{
			"index": 1,
		},
	})

	if err != nil {
		return nil, err
	}

	index := -1
	count := 0
	for i, d := range devices {
		count++
		if d.Index != i {
			index = i
			break
		}
	}
	if index == -1 {
		index = count
	}
	ipSplits := strings.Split(clusterAccount.WgIp, ".")
	ip := fmt.Sprintf("10.12.%v.%v", ipSplits[2], index+2)
	fmt.Println(ip)
	newDevice, e := d.deviceRepo.Create(ctx, &entities.Device{
		Name:       deviceName,
		ClusterId:  clusterId,
		AccountId:  accountId,
		UserId:     userId,
		PrivateKey: &pkString,
		PublicKey:  &pbKeyString,
		Ip:         ip,
		Status:     entities.DeviceStateSyncing,
		Index:      index,
	})

	if e != nil {
		return nil, fmt.Errorf("unable to persist in db %v", e)
	}

	//e = SendAction(d.infraMessenger, entities.AddPeerAction{
	//	ClusterID: clusterId,
	//	PublicKey: pbKeyString,
	//	PeerIp:    ip,
	//	AccountId: string(accountId),
	//})
	if e != nil {
		return nil, e
	}

	return newDevice, e
}

func (d *domain) RemoveDevice(ctx context.Context, deviceId repos.ID) error {
	device, err := d.deviceRepo.FindById(ctx, deviceId)
	if err != nil {
		return err
	}
	device.Status = entities.DeviceStateSyncing
	_, err = d.deviceRepo.UpdateById(ctx, deviceId, device)
	if err != nil {
		return err
	}
	//err = SendAction(d.infraMessenger, entities.DeletePeerAction{
	//	ClusterID: device.ClusterId,
	//	DeviceID:  device.Id,
	//	PublicKey: *device.PublicKey,
	//})
	if err != nil {
		return err
	}
	return err
}

func (d *domain) ListClusterDevices(ctx context.Context, clusterId *repos.ID, accountId *repos.ID) ([]*entities.Device, error) {
	q := make(repos.Filter)
	if clusterId == nil {
		q["cluster_id"] = *clusterId
	}
	if accountId != nil {
		q["account_id"] = *accountId
	}
	return d.deviceRepo.Find(ctx, repos.Query{
		Filter: q,
	})
}

func (d *domain) ListUserDevices(ctx context.Context, userId repos.ID, clusterId *repos.ID, accountId *repos.ID) ([]*entities.Device, error) {
	q := make(repos.Filter)
	if clusterId != nil {
		q["cluster_id"] = *clusterId
	}
	if accountId != nil {
		q["account_id"] = *accountId
	}
	q["user_id"] = userId
	return d.deviceRepo.Find(ctx, repos.Query{
		Filter: q,
	})
}

func (d *domain) GetCluster(ctx context.Context, id repos.ID) (*entities.Cluster, error) {
	return d.clusterRepo.FindById(ctx, id)
}

func (d *domain) GetClusters(ctx context.Context) ([]*entities.Cluster, error) {
	return d.clusterRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{},
	})
}

func (d *domain) GetDevice(ctx context.Context, id repos.ID) (*entities.Device, error) {
	return d.deviceRepo.FindById(ctx, id)
}

type Env struct {
	KafkaInfraTopic      string `env:"KAFKA_INFRA_TOPIC" required:"true"`
	ManagedTemplatesPath string `env:"MANAGED_TEMPLATES_PATH" required:"true"`
	ComputePlansPath     string `env:"COMPUTE_PLANS_PATH" required:"true"`
}

func fxDomain(
	notifier rcn.ResourceChangeNotifier,
	deviceRepo repos.DbRepo[*entities.Device],
	clusterRepo repos.DbRepo[*entities.Cluster],
	projectRepo repos.DbRepo[*entities.Project],
	configRepo repos.DbRepo[*entities.Config],
	secretRepo repos.DbRepo[*entities.Secret],
	routerRepo repos.DbRepo[*entities.Router],
	appRepo repos.DbRepo[*entities.App],
	managedSvcRepo repos.DbRepo[*entities.ManagedService],
	managedResRepo repos.DbRepo[*entities.ManagedResource],
	clusterAccountRepo repos.DbRepo[*entities.ClusterAccount],
	msgP redpanda.Producer,
	env *Env,
	logger logger.Logger,
	workloadMessenger WorkloadMessenger,
	ciClient ci.CIClient,
	iamClient iam.IAMClient,
	authClient auth.AuthClient,
	financeClient finance.FinanceClient,
	changeNotifier rcn.ResourceChangeNotifier,
) Domain {
	return &domain{
		clusterAccountRepo:   clusterAccountRepo,
		changeNotifier:       changeNotifier,
		notifier:             notifier,
		ciClient:             ciClient,
		authClient:           authClient,
		iamClient:            iamClient,
		workloadMessenger:    workloadMessenger,
		deviceRepo:           deviceRepo,
		clusterRepo:          clusterRepo,
		projectRepo:          projectRepo,
		routerRepo:           routerRepo,
		secretRepo:           secretRepo,
		configRepo:           configRepo,
		appRepo:              appRepo,
		managedSvcRepo:       managedSvcRepo,
		managedResRepo:       managedResRepo,
		messageProducer:      msgP,
		messageTopic:         env.KafkaInfraTopic,
		managedTemplatesPath: env.ManagedTemplatesPath,
		logger:               logger,
		financeClient:        financeClient,
		computePlansPath:     env.ComputePlansPath,
	}
}

var Module = fx.Module(
	"domain",
	config.EnvFx[Env](),
	fx.Provide(fxDomain),
)
