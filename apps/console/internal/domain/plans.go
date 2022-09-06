package domain

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/errors"
)

func (d *domain) GetStoragePlans(_ context.Context) ([]entities.StoragePlan, error) {
	fileData, err := ioutil.ReadFile(fmt.Sprintf("%s/storage-plans.yaml", d.inventoryPath))
	if err != nil {
		return nil, err
	}

	var plans []entities.StoragePlan

	err = yaml.Unmarshal(fileData, &plans)
	if err != nil {
		return nil, err
	}

	return plans, nil
}

func (d *domain) GetComputePlan(_ context.Context, name string) (*entities.ComputePlan, error) {
	return d.getComputePlan(name)
}

func (d *domain) GetComputePlans(_ context.Context) ([]entities.ComputePlan, error) {
	fileData, err := ioutil.ReadFile(fmt.Sprintf("%s/compute-plans.yaml", d.inventoryPath))
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

func (d *domain) getStoragePlan(name string) (*entities.StoragePlan, error) {
	fileData, err := ioutil.ReadFile(fmt.Sprintf("%s/storage-plans.yaml", d.inventoryPath))
	if err != nil {
		return nil, err
	}

	var plans []entities.StoragePlan
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

func (d *domain) getComputePlan(name string) (*entities.ComputePlan, error) {
	fileData, err := ioutil.ReadFile(fmt.Sprintf("%s/compute-plans.yaml", d.inventoryPath))
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
