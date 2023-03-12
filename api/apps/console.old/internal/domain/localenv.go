package domain

import (
	"context"
	"fmt"
	"kloudlite.io/apps/console.old/internal/domain/entities/localenv"
	"kloudlite.io/pkg/repos"
	"strings"
)

func (d *domain) GenerateEnv(ctx context.Context, klfile localenv.KLFile) (map[string]string, map[string]string, error) {
	envVars := map[string]string{}
	mountFiles := map[string]string{}
	for _, resource := range klfile.Configs {
		c, err := d.configRepo.FindOne(ctx, repos.Filter{
			"name": resource.Name,
		})
		if err != nil {
			return nil, nil, err
		}
		cmap := map[string]string{}
		for _, entry := range c.Data {
			cmap[entry.Key] = entry.Value
		}
		for _, e := range resource.Env {
			envVars[e.Key] = cmap[e.RefKey]
		}
	}
	for _, resource := range klfile.Secrets {
		c, err := d.secretRepo.FindOne(ctx, repos.Filter{
			"name": resource.Name,
		})
		if err != nil {
			return nil, nil, err
		}
		cmap := map[string]string{}
		for _, entry := range c.Data {
			cmap[entry.Key] = entry.Value
		}
		for _, e := range resource.Env {
			envVars[e.Key] = cmap[e.RefKey]
		}
	}
	for _, e := range klfile.Env {
		envVars[e.Key] = e.Value
	}
	for _, resource := range klfile.Mres {
		splits := strings.Split(resource.Name, "/")
		if len(splits) != 2 {
			return nil, nil, fmt.Errorf("invalid managed resource format: %s", resource.Name)
		}
		managedSvc, err := d.managedSvcRepo.FindOne(ctx, repos.Filter{
			"name": splits[0],
		})
		if err != nil {
			return nil, nil, err
		}
		mres, err := d.managedResRepo.FindOne(ctx, repos.Filter{
			"name":       splits[1],
			"service_id": managedSvc.Id,
		})
		if err != nil {
			return nil, nil, err
		}
		outputs, err := d.GetManagedResOutput(ctx, mres.Id)
		if err != nil {
			return nil, nil, err
		}
		for _, e := range resource.Env {
			envVars[e.Key] = outputs[e.RefKey].(string)
		}
	}
	for _, mount := range klfile.FileMount.Mounts {
		if mount.Type == "config" {
			config, err := d.configRepo.FindOne(ctx, repos.Filter{
				"name": mount.Name,
			})
			if err != nil {
				return nil, nil, err
			}
			for _, e := range config.Data {
				mountFiles[fmt.Sprintf("%v/%v", mount.Path, e.Key)] = e.Value
			}
		}
		if mount.Type == "secret" {
			secret, err := d.secretRepo.FindOne(ctx, repos.Filter{
				"name": mount.Name,
			})
			if err != nil {
				return nil, nil, err
			}
			for _, e := range secret.Data {
				mountFiles[fmt.Sprintf("%v/%v", mount.Path, e.Key)] = e.Value
			}
		}
	}
	return envVars, mountFiles, nil
}
