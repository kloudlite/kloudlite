package app

import (
	"context"
	_ "fmt"
	"kloudlite.io/common"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/logger"
	"net/http"
	_ "net/http"

	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/config"
	_ "kloudlite.io/pkg/logger"

	_ "github.com/99designs/gqlgen/graphql/handler"
	_ "github.com/99designs/gqlgen/graphql/playground"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
	"kloudlite.io/apps/console/internal/app/graph"
	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/domain"
	"kloudlite.io/apps/console/internal/domain/entities"
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
	fx.Provide(func(db *mongo.Database) (
		repos.DbRepo[*entities.Cluster],
		repos.DbRepo[*entities.Device],
	) {
		deviceRepo := repos.NewMongoRepoAdapter[*entities.Device](
			db,
			"devices",
			"dev",
			repos.MongoRepoOptions{
				IndexFields: []string{"id", "name", "cluster_id", "user_id", "ip", "index"},
			},
		)
		clusterRepo := repos.NewMongoRepoAdapter[*entities.Cluster](db, "clusters", "cls",
			repos.MongoRepoOptions{
				IndexFields: []string{"id", "name", "provider", "region", "ip", "index", "status"},
			})
		return clusterRepo, deviceRepo
	}),

	fx.Invoke(func(
		lifecycle fx.Lifecycle,
		dr repos.DbRepo[*entities.Cluster],
		cr repos.DbRepo[*entities.Device],
	) {
		lifecycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				err := dr.IndexFields(ctx)
				if err != nil {
					return err
				}
				return cr.IndexFields(ctx)
				return nil
			},
		})
	}),

	fx.Module("producer",
		fx.Provide(func(messagingCli messaging.KafkaClient) (messaging.Producer[messaging.Json], error) {
			return messaging.NewKafkaProducer[messaging.Json](messagingCli)
		}),
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
	),

	domain.Module,

	fx.Invoke(func(
		server *http.ServeMux,
		d domain.Domain,
		cacheClient cache.Client,
		logger logger.Logger,
	) {
		schema := generated.NewExecutableSchema(
			generated.Config{Resolvers: &graph.Resolver{Domain: d}},
		)
		httpServer.SetupGQLServer(
			server,
			schema,
			cache.NewSessionRepo[*common.AuthSession](
				cacheClient,
				"hotspot-session",
				"hotspot:auth:sessions",
			),
		)
	}),
)
