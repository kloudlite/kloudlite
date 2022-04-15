package domain

import (
	"context"

	"go.uber.org/fx"
	"kloudlite.io/apps/finance/internal/domain/entities"
	"kloudlite.io/pkg/config"
	"kloudlite.io/pkg/logger"
	"kloudlite.io/pkg/messaging"
	"kloudlite.io/pkg/repos"
)

type domain struct {
	deviceRepo      repos.DbRepo[*entities.Device]
	clusterRepo     repos.DbRepo[*entities.Cluster]
	messageProducer messaging.Producer[messaging.Json]
	messageTopic    string
	logger          logger.Logger
	infraMessenger  InfraMessenger
}

// Activate implements Domain
func (*domain) Activate() {
	panic("unimplemented")
}

// AddMemberships implements Domain
func (*domain) AddMemberships() {
	panic("unimplemented")
}

// CreateAccount implements Domain
func (*domain) CreateAccount() {
	panic("unimplemented")
}

// Deactivate implements Domain
func (*domain) Deactivate() {
	panic("unimplemented")
}

// DeleteAccount implements Domain
func (*domain) DeleteAccount() {
	panic("unimplemented")
}

// EnsureAccount implements Domain
func (*domain) EnsureAccount() {
	panic("unimplemented")
}

// GetAccount implements Domain
func (*domain) GetAccount() {
	panic("unimplemented")
}

// GetStripeSetupIntent implements Domain
func (*domain) GetStripeSetupIntent() {
	panic("unimplemented")
}

// InviteMember implements Domain
func (*domain) InviteMember() {
	panic("unimplemented")
}

// ListAccount implements Domain
func (*domain) ListAccount() {
	panic("unimplemented")
}

// ListMemberships implements Domain
func (*domain) ListMemberships() {
	panic("unimplemented")
}

// RemoveMember implements Domain
func (*domain) RemoveMember() {
	panic("unimplemented")
}

// RemoveMemberships implements Domain
func (*domain) RemoveMemberships() {
	panic("unimplemented")
}

// UpdateAccount implements Domain
func (*domain) UpdateAccount() {
	panic("unimplemented")
}

// UpdateBilling implements Domain
func (*domain) UpdateBilling() {
	panic("unimplemented")
}

// UpdateMember implements Domain
func (*domain) UpdateMember() {
	panic("unimplemented")
}

func (d *domain) GetCluster(ctx context.Context, id repos.ID) (*entities.Cluster, error) {
	return d.clusterRepo.FindById(ctx, id)
}

func (d *domain) GetDevice(ctx context.Context, id repos.ID) (*entities.Device, error) {
	return d.deviceRepo.FindById(ctx, id)
}

type Env struct {
	KafkaInfraTopic string `env:"KAFKA_INFRA_TOPIC" required:"true"`
}

func fxDomain(
	deviceRepo repos.DbRepo[*entities.Device],
	clusterRepo repos.DbRepo[*entities.Cluster],
	msgP messaging.Producer[messaging.Json],
	env *Env,
	logger logger.Logger,
	messenger InfraMessenger,
) Domain {
	return &domain{
		infraMessenger:  messenger,
		deviceRepo:      deviceRepo,
		clusterRepo:     clusterRepo,
		messageProducer: msgP,
		messageTopic:    env.KafkaInfraTopic,
		logger:          logger,
	}
}

var Module = fx.Module(
	"domain",
	fx.Provide(config.LoadEnv[Env]()),
	fx.Provide(fxDomain),
)
