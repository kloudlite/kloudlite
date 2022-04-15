package app

import (
	"context"

	"go.uber.org/fx"
	"kloudlite.io/apps/agent/internal/domain"
	"kloudlite.io/pkg/shared"
)

var projectRes = domain.Message{
	ResourceType: shared.RESOURCE_PROJECT,
	Namespace:    "hotspot",
	Spec: domain.Project{
		Name:        "sample-xyz",
		DisplayName: "this is not just a project",
		Logo:        "i have no logo",
	},
}

var managedSvcMsg = domain.Message{
	ResourceType: shared.RESOURCE_MANAGED_SERVICE,
	Namespace:    "hotspot",
	Spec: domain.ManagedSvc{
		Name:         "sample-xyz",
		Namespace:    "hotspot",
		TemplateName: "msvc_mongo",
		Version:      1,
		Values: map[string]interface{}{
			"hi": "asdfa",
		},
		LastApplied: M{"hello": "world", "something": map[string]interface{}{
			"one": 2,
			"two": 2,
		}},
	},
}

var managedApp = domain.Message{
	ResourceType: shared.RESOURCE_APP,
	Namespace:    "hotspot",
	Spec: domain.App{
		Name:      "sample",
		Namespace: "hotspot",
		Services: []domain.AppSvc{
			{
				Port:       21323,
				TargetPort: 21345,
				Type:       "tcp",
			},
		},
		Containers: []domain.AppContainer{
			{
				Name:            "sample",
				Image:           "nginx",
				ImagePullPolicy: "Always",
				Command:         []string{"hello", "world"},
				ResourceCpu:     domain.ContainerResource{Min: "100", Max: "200"},
				ResourceMemory:  domain.ContainerResource{Min: "200", Max: "300"},
				Env: []domain.ContainerEnv{
					{
						Key:   "hello",
						Value: "world",
					},
				},
			},
		},
	},
}

var managedRes = domain.Message{
	ResourceType: shared.RESOURCE_MANAGED_RESOURCE,
	Namespace:    "hotspot",
	Spec: domain.ManagedRes{
		Name:       "sample-mres",
		Type:       "db",
		Namespace:  "hotspot",
		ManagedSvc: "sample1234",
		Values: map[string]interface{}{
			"hello":  "world",
			"sample": "hello",
		},
	},
}

var configRes = domain.Message{
	ResourceType: shared.RESOURCE_CONFIG,
	Namespace:    "hotspot",
	Spec: domain.Config{
		Name:      "hi-config",
		Namespace: "hotspot",
		Data: map[string]interface{}{
			"hi":  "hello there",
			"one": 2,
		},
	},
}

var secretRes = domain.Message{
	ResourceType: shared.RESOURCE_SECRET,
	Namespace:    "hotspot",
	Spec: domain.Secret{
		Name:      "hi-config",
		Namespace: "hotspot",
		Data: map[string]interface{}{
			"hi":  "hello there",
			"one": 2,
		},
	},
}

var routerRes = domain.Message{
	ResourceType: shared.RESOURCE_ROUTER,
	Namespace:    "hotspot",
	Spec: domain.Router{
		Name:      "sample-router",
		Namespace: "hotspot",
		Domains:   []string{"x.kloudlite.io", "y.kloudlitle.io"},
		Routes: []domain.Routes{
			domain.Routes{
				Path: "/",
				App:  "sample",
				Port: 80,
			},
			domain.Routes{
				Path: "/api",
				App:  "sample-api",
				Port: 3000,
			},
		},
	},
}

var pipelineRes = domain.Message{
	ResourceType: shared.RESOURCE_GIT_PIPELINE,
	Namespace:    "hotspot",
	Spec: domain.Pipeline{
		Name:        "sample-p",
		Namespace:   "hotspot",
		GitProvider: "gitlab",
		GitRepoUrl:  "https://gitlab.com/madhouselabs/kloudlite/api-go",
		GitRef:      "heads/feature/ci",
		BuildArgs: []domain.BuildArg{
			domain.BuildArg{
				Key:   "app",
				Value: "message-consumer",
			},
		},
		// Github:     domain.PipelineGithub{},
		// Gitlab:     domain.PipelineGitlab{},
	},
}

var TModule = fx.Module("app.trial",
	fx.Invoke(func(lf fx.Lifecycle, d domain.Domain) {
		lf.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return d.ProcessMessage(ctx, &projectRes)
			},
		})
	}),
)
