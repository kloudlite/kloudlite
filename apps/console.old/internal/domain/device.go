package domain

import (
	"context"
	"errors"
	"fmt"

	op_crds "kloudlite.io/apps/console.old/internal/domain/op-crds"
	"kloudlite.io/constants"
	"kloudlite.io/pkg/beacon"
	"kloudlite.io/pkg/kubeapi"

	"kloudlite.io/apps/console.old/internal/domain/entities"
	internal_crds "kloudlite.io/apps/console.old/internal/domain/op-crds/internal-crds"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetDevice(ctx context.Context, id repos.ID) (*entities.Device, error) {
	userId, err := GetUserId(ctx)
	if err != nil {
		return nil, err
	}
	dev, err := d.deviceRepo.FindById(ctx, id)
	if err = mongoError(err, "device not found"); err != nil {
		return nil, err
	}
	if dev.UserId != repos.ID(userId) {
		return nil, errors.New("you don't have to access this resource")
	}
	return dev, nil
}

func (d *domain) ListAccountDevices(ctx context.Context, accountId repos.ID) ([]*entities.Device, error) {
	err := d.checkAccountAccess(ctx, accountId, ReadAccount)
	if err != nil {
		return nil, err
	}
	q := make(repos.Filter)
	q["account_id"] = accountId
	return d.deviceRepo.Find(
		ctx, repos.Query{
			Filter: q,
		},
	)
}

func (d *domain) ListUserDevices(ctx context.Context, userId repos.ID) ([]*entities.Device, error) {
	q := make(repos.Filter)
	q["user_id"] = userId
	return d.deviceRepo.Find(
		ctx, repos.Query{
			Filter: q,
		},
	)
}

func (d *domain) GetDeviceConfig(ctx context.Context, deviceId repos.ID) (map[string]any, error) {
	dev, err := d.deviceRepo.FindById(ctx, deviceId)
	if err != nil {
		return nil, err
	}
	cluster, err := d.getClusterForAccount(ctx, dev.AccountId)
	if err != nil {
		return nil, err
	}
	kubecli := kubeapi.NewClientWithConfigPath(fmt.Sprintf("%s/%s", d.clusterConfigsPath, getClusterKubeConfig(cluster)))

	secret, err := kubecli.GetSecret(ctx, fmt.Sprint("wg-", dev.AccountId), fmt.Sprint("wg-device-config-", dev.Id))
	if err != nil {
		return nil, err
	}
	parsedSec := make(map[string]any)
	for k, v := range secret.Data {
		parsedSec[k] = string(v)
	}
	return parsedSec, nil
}

func (d *domain) DeviceByNameExists(ctx context.Context, accountId repos.ID, name string) (bool, error) {
	one, err := d.deviceRepo.FindOne(
		ctx, repos.Filter{
			"account_id": accountId,
			"name":       name,
		},
	)
	if err != nil {
		return false, err
	}
	return one != nil, nil
}

func (d *domain) AddDevice(ctx context.Context, deviceName string, accountId repos.ID, userId repos.ID) (*entities.Device, error) {
	devices, err := d.deviceRepo.Find(
		ctx, repos.Query{
			Filter: repos.Filter{
				"account_id": accountId,
			},
			Sort: map[string]any{
				"index": 1,
			},
		},
	)
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
	device, e := d.deviceRepo.Create(
		ctx, &entities.Device{
			Name:      deviceName,
			AccountId: accountId,
			UserId:    userId,
			Status:    entities.DeviceStateSyncing,
			Index:     index,
		},
	)
	if e != nil {
		return nil, fmt.Errorf("unable to persist in db %v", e)
	}

	clusterId, err := d.getClusterForAccount(ctx, accountId)
	if err != nil {
		return nil, err
	}

	if err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(device.Id), &internal_crds.Device{
			APIVersion: internal_crds.DeviceAPIVersion,
			Kind:       internal_crds.DeviceKind,
			Metadata: internal_crds.DeviceMetadata{
				Name: string(device.Id),
				Annotations: map[string]string{
					"kloudlite.io/account-id":   string(device.AccountId),
					"kloudlite.io/resource-ref": string(device.Id),
				},
				Labels: map[string]string{
					"kloudlite.io/account-id": string(device.AccountId),
					"kloudlite.io/device-id":  string(device.Id),
				},
			},
			Spec: internal_crds.DeviceSpec{
				Account: string(device.AccountId),
				ActiveRegion: func() string {
					if device.ActiveRegion != nil {
						return *device.ActiveRegion
					}
					return ""
				}(),
				DeviceName: deviceName,
				Offset:     device.Index,
				DeviceId:   string(device.Id),
				Ports: func() []internal_crds.Port {
					var p []internal_crds.Port
					for _, p2 := range device.ExposedPorts {
						p = append(
							p, internal_crds.Port{
								Port:       p2.Port,
								TargetPort: p2.TargetPort,
							},
						)
					}
					return p
					// device.ExposedPorts
				}(),
			},
		},
	); err != nil {
		return device, err
	}

	d.beacon.TriggerWithUserCtx(ctx, accountId, beacon.EventAction{
		Action:       constants.CreateDevice,
		Status:       beacon.StatusOK(),
		ResourceType: constants.ResourceDevice,
		ResourceId:   device.Id,
	})

	return device, e
}

func (d *domain) OnDeleteDevice(ctx context.Context, response *op_crds.StatusUpdate) error {
	return d.deviceRepo.DeleteById(ctx, repos.ID(response.Metadata.ResourceId))
}

func (d *domain) OnUpdateDevice(ctx context.Context, response *op_crds.StatusUpdate) error {
	one, err := d.deviceRepo.FindById(ctx, repos.ID(response.Metadata.ResourceId))
	if err = mongoError(err, "managed resource not found"); err != nil {
		// Ignore unknown project
		return nil
	}

	if response.IsReady {
		one.Status = entities.DeviceStateAttached
	} else {
		one.Status = entities.DeviceStateSyncing
	}
	one.Conditions = response.ChildConditions
	_, err = d.deviceRepo.UpdateById(ctx, one.Id, one)
	return err
}

func (d *domain) RemoveDevice(ctx context.Context, deviceId repos.ID) error {
	device, err := d.deviceRepo.FindById(ctx, deviceId)
	if err != nil {
		return err
	}
	device.Status = entities.DeviceStateDeleted
	_, err = d.deviceRepo.UpdateById(ctx, deviceId, device)
	if err != nil {
		return err
	}

	clusterId, err := d.getClusterForAccount(ctx, device.AccountId)

	if err = d.workloadMessenger.SendAction(
		"delete", d.getDispatchKafkaTopic(clusterId), string(device.Id), &internal_crds.Device{
			APIVersion: internal_crds.DeviceAPIVersion,
			Kind:       internal_crds.DeviceKind,
			Metadata: internal_crds.DeviceMetadata{
				Name: string(device.Id),
			},
		},
	); err != nil {
		return err
	}

	go d.beacon.TriggerWithUserCtx(ctx, device.AccountId, beacon.EventAction{
		Action:       constants.DeleteDevice,
		Status:       beacon.StatusOK(),
		ResourceType: constants.ResourceDevice,
		ResourceId:   device.Id,
	})

	return nil
}

func (d *domain) UpdateDevice(ctx context.Context, deviceId repos.ID, deviceName *string, region *string, ports []entities.Port) (done bool, e error) {
	device, e := d.deviceRepo.FindById(ctx, deviceId)
	if region != nil {
		device.ActiveRegion = region
	}
	if deviceName != nil {
		device.Name = *deviceName
	}
	if ports != nil {
		device.ExposedPorts = func() []entities.Port {
			p := []entities.Port{}
			for _, p2 := range ports {
				p = append(
					p, entities.Port{
						Port:       p2.Port,
						TargetPort: p2.TargetPort,
					},
				)
			}

			return p
		}()
	}
	_, err := d.deviceRepo.UpdateById(ctx, deviceId, device)
	if err != nil {
		return false, e
	}

	clusterId, err := d.getClusterForAccount(ctx, device.AccountId)
	if err != nil {
		return false, err
	}

	if err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(device.Id), &internal_crds.Device{
			APIVersion: internal_crds.DeviceAPIVersion,
			Kind:       internal_crds.DeviceKind,
			Metadata: internal_crds.DeviceMetadata{
				Name: string(device.Id),
				Annotations: map[string]string{
					"kloudlite.io/account-id":   string(device.AccountId),
					"kloudlite.io/resource-ref": string(device.Id),
				},
				Labels: map[string]string{
					"kloudlite.io/account-id": string(device.AccountId),
					"kloudlite.io/device-id":  string(device.Id),
				},
			},
			Spec: internal_crds.DeviceSpec{
				DeviceName: device.Name,
				Account:    string(device.AccountId),
				ActiveRegion: func() string {
					if device.ActiveRegion != nil {
						return *device.ActiveRegion
					}
					return ""
				}(),
				Offset:   device.Index,
				DeviceId: string(device.Id),
				Ports: func() []internal_crds.Port {
					p := make([]internal_crds.Port, 0)
					for _, p2 := range device.ExposedPorts {
						p = append(
							p, internal_crds.Port{
								Port:       p2.Port,
								TargetPort: p2.TargetPort,
							},
						)
					}
					return p
				}(),
			},
		},
	); err != nil {
		return false, err
	}

	go d.beacon.TriggerWithUserCtx(ctx, device.AccountId, beacon.EventAction{
		Action:       constants.UpdateDevice,
		Status:       beacon.StatusOK(),
		ResourceType: constants.ResourceDevice,
		ResourceId:   device.Id,
	})

	return true, nil
}
