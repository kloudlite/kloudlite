package domain

import (
	"context"
	"go.uber.org/fx"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	GetPipeline(ctx context.Context, pipelineId repos.ID) (*Pipeline, error)
	CretePipeline(ctx context.Context, pipeline Pipeline) (*Pipeline, error)
}

type domainI struct {
	pipelineRepo repos.DbRepo[*Pipeline]
}

func (d domainI) CretePipeline(ctx context.Context, pipeline Pipeline) (*Pipeline, error) {
	return d.pipelineRepo.Create(ctx, &pipeline)
}

func (d domainI) GetPipeline(ctx context.Context, pipelineId repos.ID) (*Pipeline, error) {
	id, err := d.pipelineRepo.FindById(ctx, pipelineId)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func fxDomain(pipelineRepo repos.DbRepo[*Pipeline]) Domain {
	return domainI{
		pipelineRepo: pipelineRepo,
	}
}

var Module = fx.Module("domain",
	fx.Provide(fxDomain),
)
