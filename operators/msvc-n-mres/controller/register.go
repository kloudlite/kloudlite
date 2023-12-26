package controller

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	elasticsearchMsvcv1 "github.com/kloudlite/operator/apis/elasticsearch.msvc/v1"
	influxdbMsvcv1 "github.com/kloudlite/operator/apis/influxdb.msvc/v1"
	mongodbMsvcv1 "github.com/kloudlite/operator/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "github.com/kloudlite/operator/apis/mysql.msvc/v1"
	neo4jMsvcv1 "github.com/kloudlite/operator/apis/neo4j.msvc/v1"
	redisMsvcv1 "github.com/kloudlite/operator/apis/redis.msvc/v1"
	redpandaMsvcv1 "github.com/kloudlite/operator/apis/redpanda.msvc/v1"
	zookeeperMsvcv1 "github.com/kloudlite/operator/apis/zookeeper.msvc/v1"
	"github.com/kloudlite/operator/operator"
	cmsvc "github.com/kloudlite/operator/operators/msvc-n-mres/internal/cluster-msvc"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/env"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/mres"
	"github.com/kloudlite/operator/operators/msvc-n-mres/internal/msvc"
)

func RegisterInto(mgr operator.Operator) {
	ev := env.GetEnvOrDie()
	mgr.AddToSchemes(
		crdsv1.AddToScheme,
		mongodbMsvcv1.AddToScheme,
		mysqlMsvcv1.AddToScheme,
		redisMsvcv1.AddToScheme,
		elasticsearchMsvcv1.AddToScheme,
		influxdbMsvcv1.AddToScheme,
		redpandaMsvcv1.AddToScheme,
		zookeeperMsvcv1.AddToScheme,
		neo4jMsvcv1.AddToScheme,
	)
	mgr.RegisterControllers(
		&cmsvc.Reconciler{Name: "cmsvc", Env: ev},
		&msvc.Reconciler{Name: "msvc", Env: ev},
		&mres.Reconciler{Name: "mres", Env: ev},
	)
}
