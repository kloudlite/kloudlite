package domain

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/pkg/errors"
	httpServer "kloudlite.io/pkg/http-server"
	"kloudlite.io/pkg/repos"
)

const (
	ReadProject   = "read_project"
	UpdateProject = "update_project"
	ReadAccount   = "read_account"
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

func (d *domain) getDispatchKafkaTopic(clusterId string) string {
	return clusterId + "-incoming"
}

func (d *domain) getClusterIdForProject(ctx context.Context, projectId repos.ID) (string, error) {
	project, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return "", err
	}

	clusterId, err := d.getClusterForAccount(ctx, project.AccountId)
	if err != nil {
		return "", err
	}
	return clusterId, nil
}
