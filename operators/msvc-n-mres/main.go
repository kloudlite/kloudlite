package main

import (
	crdsv1 "operators.kloudlite.io/apis/crds/v1"
	elasticsearchMsvcv1 "operators.kloudlite.io/apis/elasticsearch.msvc/v1"
	influxdbMsvcv1 "operators.kloudlite.io/apis/influxdb.msvc/v1"
	mongodbMsvcv1 "operators.kloudlite.io/apis/mongodb.msvc/v1"
	mysqlMsvcv1 "operators.kloudlite.io/apis/mysql.msvc/v1"
	neo4jMsvcv1 "operators.kloudlite.io/apis/neo4j.msvc/v1"
	redisMsvcv1 "operators.kloudlite.io/apis/redis.msvc/v1"
	redpandaMsvcv1 "operators.kloudlite.io/apis/redpanda.msvc/v1"
	zookeeperMsvcv1 "operators.kloudlite.io/apis/zookeeper.msvc/v1"
	"operators.kloudlite.io/operator"
	"operators.kloudlite.io/operators/msvc-n-mres/internal/env"
	"operators.kloudlite.io/operators/msvc-n-mres/internal/mres"
	"operators.kloudlite.io/operators/msvc-n-mres/internal/msvc"
)

func main() {
	ev := env.GetEnvOrDie()
	mgr := operator.New("msvc-and-mres")
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
		&msvc.ManagedServiceReconciler{Name: "msvc", Env: ev},
		&mres.ManagedResourceReconciler{Name: "mres", Env: ev},
	)
	mgr.Start()
}
