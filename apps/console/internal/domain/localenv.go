package domain

import (
	"context"
	"fmt"
	"kloudlite.io/apps/console/internal/domain/entities/localenv"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GenerateEnv(ctx context.Context, klfile localenv.KLFile) (string, map[string]string, error) {
	var envVars map[string]string
	var mountFiles map[string]string
	for _, resource := range klfile.Configs {
		c, err := d.configRepo.FindById(ctx, resource.Id)
		if err != nil {
			return "", nil, err
		}
		var cmap map[string]string
		for _, entry := range c.Data {
			cmap[entry.Key] = entry.Value
		}
		for _, e := range resource.Env {
			envVars[e.Key] = cmap[e.RefKey]
		}
	}
	for _, resource := range klfile.Secrets {
		c, err := d.secretRepo.FindById(ctx, resource.Id)
		if err != nil {
			return "", nil, err
		}
		var cmap map[string]string
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
		outputs, err := d.GetManagedResOutput(ctx, resource.Id)
		if err != nil {
			return "", nil, err
		}
		for _, e := range resource.Env {
			envVars[e.Key] = outputs[e.RefKey].(string)
		}
	}
	for _, mount := range klfile.FileMount.Mounts {
		if mount.Type == "config" {
			config, err := d.configRepo.FindById(ctx, repos.ID(mount.Ref))
			if err != nil {
				return "", nil, err
			}
			for _, e := range config.Data {
				mountFiles[fmt.Sprintf("/configs/%v/%v", config.Name, e.Key)] = e.Value
			}
		}
		if mount.Type == "secret" {
			secret, err := d.secretRepo.FindById(ctx, repos.ID(mount.Ref))
			if err != nil {
				return "", nil, err
			}
			for _, e := range secret.Data {
				mountFiles[fmt.Sprintf("/secrets/%v/%v", secret.Name, e.Key)] = e.Value
			}
		}
	}
	envExport := ""
	for k, v := range envVars {
		envExport += fmt.Sprintf("%v=%v\n", k, v)
	}
	return envExport, mountFiles, nil
}
