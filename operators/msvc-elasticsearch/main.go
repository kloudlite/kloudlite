package main

import (
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	elasticsearchmsvcv1 "operators.kloudlite.io/apis/elasticsearch.msvc/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-elasticsearch/internal/controllers/kibana"
	"operators.kloudlite.io/operators/msvc-elasticsearch/internal/controllers/service"
	"operators.kloudlite.io/operators/msvc-elasticsearch/internal/env"
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
