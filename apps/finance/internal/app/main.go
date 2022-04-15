package app

import (
	"context"
	"fmt"
	"kloudlite.io/pkg/config"
	"net/http"

	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"kloudlite.io/apps/finance/internal/app/graph"
	"kloudlite.io/apps/finance/internal/app/graph/generated"
	"kloudlite.io/apps/finance/internal/domain"
	"kloudlite.io/apps/finance/internal/domain/entities"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/repos"
)

type Env struct {
	KafkaInfraTopic         string `env:"KAFKA_INFRA_TOPIC"`
	KafkaInfraResponseTopic string `env:"KAFKA_INFRA_RESP_TOPIC"`
	KafkaConsumerGroupId    string `env:"KAFKA_GROUP_ID"`
}

var Module = fx.Module(
	"app",
	fx.Provide(config.LoadEnv[Env]()),
	// Create Repos
	fx.Provide(func(db *mongo.Database) (
		repos.DbRepo[*entities.Cluster],
		repos.DbRepo[*entities.Device],
	) {
		deviceRepo := repos.NewMongoRepoAdapter[*entities.Device](db, "devices", "dev")
		clusterRepo := repos.NewMongoRepoAdapter[*entities.Cluster](db, "clusters", "cls")
		return clusterRepo, deviceRepo
	}),

	fx.Provide(func(messagingCli messaging.KafkaClient) (messaging.Producer[messaging.Json], error) {
		return messaging.NewKafkaProducer[messaging.Json](messagingCli)
	}),

	domain.Module,

	fx.Provide(func(env *Env, p messaging.Producer[messaging.Json]) domain.InfraMessenger {
		return &infraMessengerImpl{
			env:      env,
			producer: p,
			//onAddClusterResponse: func(ctx context.Context, m entities.SetupClusterResponse) {
			//	if m.Done {
			//		d.UpdateClusterState(ctx, repos.ID(m.ClusterID), entities.ClusterStateLive)
			//		return
			//	}
			//	d.UpdateClusterState(ctx, repos.ID(m.ClusterID), entities.ClusterStateError)
			//},
			//
			//onDeleteClusterResponse: func(ctx context.Context, m entities.DeleteClusterResponse) {
			//	if m.Done {
			//		d.UpdateClusterState(ctx, repos.ID(m.ClusterID), entities.ClusterStateLive)
			//		return
			//	}
			//	d.UpdateClusterState(ctx, repos.ID(m.ClusterID), entities.ClusterStateError)
			//
			//},
			//
			//onUpdateClusterResponse: func(ctx context.Context, m entities.UpdateClusterResponse) {
			//	if m.Done {
			//		d.UpdateClusterState(ctx, repos.ID(m.ClusterID), entities.ClusterStateLive)
			//		return
			//	}
			//	d.UpdateClusterState(ctx, repos.ID(m.ClusterID), entities.ClusterStateError)
			//
			//},
			//
			//onAddDeviceResponse: func(ctx context.Context, m entities.AddPeerResponse) {
			//
			//	if m.Done {
			//		d.UpdateDeviceState(ctx, repos.ID(m.DeviceID), entities.DeviceStateAttached)
			//	}
			//},
			//onRemoveDeviceResponse: func(ctx context.Context, m entities.DeletePeerResponse) {
			//
			//},
		}
	}),

	fx.Invoke(func(producer messaging.Producer[messaging.Json], lifecycle fx.Lifecycle) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return producer.Connect(ctx)
			},
			OnStop: func(ctx context.Context) error {
				producer.Close(ctx)
				return nil
			},
		})
	}),

	fx.Invoke(func(server *http.ServeMux, d domain.Domain) {
		server.HandleFunc("/play", playground.Handler("Graphql playground", "/query"))

		gqlServer := gqlHandler.NewDefaultServer(
			generated.NewExecutableSchema(
				generated.Config{Resolvers: &graph.Resolver{Domain: d}},
			),
		)

		server.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			fmt.Printf("Headers: %+v", req.Cookies())
			gqlServer.ServeHTTP(w, req)
		})

		// server.Handle("/", gqlServer)
	}),
)
