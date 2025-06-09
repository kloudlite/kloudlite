package main

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	elasticsearchmsvcv1 "github.com/kloudlite/operator/apis/elasticsearch.msvc/v1"
	"github.com/kloudlite/operator/operator"
	"github.com/kloudlite/operator/operators/msvc-elasticsearch/internal/controllers/kibana"
	"github.com/kloudlite/operator/operators/msvc-elasticsearch/internal/controllers/service"
	"github.com/kloudlite/operator/operators/msvc-elasticsearch/internal/env"
)

func main() {
	mgr := operator.New("msvc-elasticsearch")
	mgr.AddToSchemes(
		elasticsearchmsvcv1.AddToScheme,
		crdsv1.AddToScheme,
	)
	ev := env.GetEnvOrDie()
	mgr.RegisterControllers(
		&service.Reconciler{Name: "service", Env: ev},
		&kibana.Reconciler{Name: "kibana", Env: ev},
	)
	mgr.Start()
}
