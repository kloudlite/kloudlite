package domain

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/pkg/cache"
	"kloudlite.io/pkg/errors"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"

	fWebsocket "github.com/gofiber/websocket/v2"
)

const (
	ReadProject   = "read_project"
	UpdateProject = "update_project"
	ReadAccount   = "read_account"
	UpdateAccount = "update_account"
)

func mongoError(err error, descp string) error {
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New(descp)
		}
		return err
	}
	return nil
}

func GetUser(ctx context.Context) (string, error) {

	session := httpServer.GetSession[*common.AuthSession](ctx)

	if session == nil {
		return "", errors.New("Unauthorized")
	}
	return string(session.UserId), nil
}

func (d *domain) getClusterForAccount(ctx context.Context, accountId repos.ID) (string, error) {
	cluster, err := d.financeClient.GetAttachedCluster(ctx, &finance.GetAttachedClusterIn{AccountId: string(accountId)})
	if err != nil {
		return "", errors.NewEf(err, "failed to get cluster from accountId [grpc]")
	}
	return cluster.ClusterId, nil
}

type DispatchKafkaTopicType string

const (
	SendToAgent              DispatchKafkaTopicType = "send-to-agent"
	StatusUpdatesFromAgent   DispatchKafkaTopicType = "status-updates-from-agent"
	PipelineUpdatesFromAgent DispatchKafkaTopicType = "pipeline-updates-from-agent"
)

func getClusterKubeConfig(clusterName string) string {
	return clusterName + "-kubeconfig"
}

func (d *domain) getDispatchKafkaTopic(clusterId string) string {
	return clusterId + "-incoming"
}

// func (d *domain) getClusterIdForProject(ctx context.Context, projectId repos.ID) (string, error) {
// 	project, err := d.projectRepo.FindById(ctx, projectId)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	clusterId, err := d.getClusterForAccount(ctx, project.AccountId)
// 	if err != nil {
// 		return "", err
// 	}
// 	return clusterId, nil
// }

func (*domain) GetSocketCtx(
	conn *fWebsocket.Conn,
	cacheClient AuthCacheClient,
	cookieName,
	cookieDomain string,
	sessionKeyPrefix string,
) context.Context {
	repo := cache.NewRepo[*common.AuthSession](cacheClient)
	cookieValue := conn.Cookies(cookieName)
	c := context.TODO()

	if cookieValue != "" {
		key := fmt.Sprintf("%s:%s", sessionKeyPrefix, cookieValue)
		var get any
		get, err := repo.Get(c, key)
		if err != nil {
			if !repo.ErrNoRecord(err) {
				return c
			}
		}

		fmt.Println(get)

		// if get != nil {
		// 	c = context.WithValue(c, userContextKey, context.WithValue(c, "session", get))
		// }
	}

	return c
}
