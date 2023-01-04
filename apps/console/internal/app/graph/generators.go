package graph

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/jseval"
	"kloudlite.io/pkg/repos"
)

func mountsGenerator(ctx context.Context, c model.AppContainer, d domain.Domain) ([]op_crds.Volume, error) {
	if c.Mounts == nil {
		return []op_crds.Volume{}, nil
	}
	if len(c.Mounts) == 0 {
		return []op_crds.Volume{}, nil
	}
	vs := make([]op_crds.Volume, 0)

	for _, v := range c.Mounts {
		vs = append(
			vs, op_crds.Volume{
				MountPath: v.Path,
				Type:      v.Type,
				RefName:   v.Ref,
				Items: func() []op_crds.VolumeItem {
					var items []op_crds.VolumeItem
					if v.Type == "config" {
						config, _ := d.GetConfig(ctx, repos.ID(v.Ref))
						for _, e := range config.Data {
							items = append(
								items, op_crds.VolumeItem{
									Key: e.Key,
								},
							)
						}
					} else {
						secret, _ := d.GetSecret(ctx, repos.ID(v.Ref))
						for _, e := range secret.Data {
							items = append(
								items, op_crds.VolumeItem{
									Key: e.Key,
								},
							)
						}
					}
					return items
				}(),
			},
		)
	}
	return vs, nil
}

func appSpecGenerator(ctx context.Context, app *model.App, p *entities.Project, d domain.Domain) *op_crds.AppSpec {
	return &op_crds.AppSpec{
		Region: func() string {
			if p.RegionId != nil {
				return string(*p.RegionId)
			}
			return ""
		}(),
		Services: func() []op_crds.Service {
			svcs := make([]op_crds.Service, 0)
			for _, ep := range app.Services {
				svcs = append(
					svcs, op_crds.Service{
						Port:       int(ep.Exposed),
						TargetPort: int(ep.Target),
						Type:       string(ep.Type),
					},
				)
			}
			return svcs
		}(),
		Hpa: func() *op_crds.HPA {
			if app.AutoScale == nil {
				return nil
			}
			return &op_crds.HPA{
				Enabled:     true,
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
				cs = append(
					cs, op_crds.Container{
						Name:            c.Name,
						Image:           c.Image,
						ImagePullPolicy: "Always",
						Volumes: func() []op_crds.Volume {
							cs, err := mountsGenerator(ctx, *app.Containers[0], d)
							if err != nil {
								// panic(err)
								return []op_crds.Volume{}
							}
							return cs
						}(),
						Env: func() []op_crds.EnvEntry {
							env := make([]op_crds.EnvEntry, 0)
							for _, e := range c.EnvVars {
								if e.Value.Type == "managed_service" {
									ref := fmt.Sprintf("msvc-%v", *e.Value.Ref)
									env = append(
										env, op_crds.EnvEntry{
											Value:   e.Value.Value,
											Key:     e.Key,
											Type:    "secret",
											RefName: &ref,
											RefKey:  e.Value.Key,
										},
									)
								} else if e.Value.Type == "managed_resource" {
									ref := fmt.Sprintf("mres-%v", *e.Value.Ref)
									env = append(
										env, op_crds.EnvEntry{
											Value:   e.Value.Value,
											Key:     e.Key,
											Type:    "secret",
											RefName: &ref,
											RefKey:  e.Value.Key,
										},
									)
								} else {
									env = append(
										env, op_crds.EnvEntry{
											Value:   e.Value.Value,
											Key:     e.Key,
											Type:    e.Value.Type,
											RefName: e.Value.Ref,
											RefKey:  e.Value.Key,
										},
									)
								}
							}
							return env
						}(),
						ResourceCpu: func() *op_crds.Limit {
							o := op_crds.Limit{
								Min: fmt.Sprintf(
									"%vm", int(
										c.Quantity*(func() float64 {
											if *c.IsShared {
												return 250
											}
											return 1000
										})(),
									),
								),
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
					},
				)
			}
			return cs
		}(),
		Replicas: *app.Replicas,
	}
}

func managedSvcSpecGenerator(ctx context.Context, managedSvc *model.ManagedSvc, p *entities.Project, d domain.Domain) (*op_crds.ManagedServiceSpec, error) {

	template, err := d.GetManagedServiceTemplate(ctx, string(managedSvc.Source))
	if err != nil {
		return nil, err
	}

	eval, err := d.JsEval(
		ctx, &jseval.EvalIn{
			Init:    template.InputMiddleware,
			FunName: "inputMiddleware",
			Inputs: func() *anypb.Any {
				marshal, _ := json.Marshal(managedSvc.Values)
				return &anypb.Any{
					TypeUrl: "",
					Value:   marshal,
				}
			}(),
		},
	)
	if err != nil {
		return nil, err
	}

	var transformedInputs struct {
		Inputs     map[string]any    `json:"inputs"`
		Annotation map[string]string `json:"annotation,omitempty"`
	}
	err = json.Unmarshal(eval.Output.Value, &transformedInputs)
	if err != nil {
		return nil, err
	}

	return &op_crds.ManagedServiceSpec{
		MsvcKind: op_crds.MsvcKind{
			APIVersion: template.ApiVersion,
			Kind:       template.Kind,
		},
		Region: string(*p.RegionId),
		Inputs: transformedInputs.Inputs,
	}, nil

}

func mResSpecGenerator(ctx context.Context, managedRes *model.ManagedRes, oMres *entities.ManagedResource, d domain.Domain) (*op_crds.ManagedResourceSpec, error) {

	msvc, err := d.GetManagedSvc(ctx, oMres.ServiceId)
	if err != nil {
		return nil, err
	}
	template, err := d.GetManagedServiceTemplate(ctx, string(msvc.ServiceType))
	if err != nil {
		return nil, err
	}

	return &op_crds.ManagedResourceSpec{
		MsvcRef: op_crds.MsvcRef{
			APIVersion: template.ApiVersion,
			Kind:       template.Kind,
			Name:       string(oMres.ServiceId),
		},
		MresKind: op_crds.MresKind{
			Kind: string(oMres.ResourceType),
		},
		Inputs: func() map[string]string {
			values := map[string]string{}
			managedRes.Values["resourceName"] = managedRes.Name
			for k, v := range managedRes.Values {
				values[k] = fmt.Sprintf("%s", v)
			}
			return values
		}(),
	}, nil
}

func secretOverrideDataGenerator(secret *model.Secret) map[string][]byte {
	data := make(map[string][]byte, 0)
	for _, d := range secret.Entries {
		// encoded := b64.StdEncoding.EncodeToString([]byte(d.Value))
		data[d.Key] = []byte(d.Value)
	}
	return data
}

func configOverrideDataGenerator(config *model.Config) map[string]string {
	data := make(map[string]string, 0)
	for _, i := range config.Entries {
		data[i.Key] = i.Value
	}
	return data
}
