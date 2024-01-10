package app

import (
	"context"
	"encoding/json"
	"github.com/kloudlite/api/apps/infra/internal/domain"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/infra"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	wireguardV1 "github.com/kloudlite/operator/apis/wireguard/v1"
)

type grpcServer struct {
	d domain.Domain
	infra.UnimplementedInfraServer
}

func (g *grpcServer) contextFromAccount(ctx context.Context, accountName string) domain.InfraContext {
	dctx := domain.InfraContext{
		Context:     ctx,
		UserId:      "sys-user:error-on-apply-worker",
		UserEmail:   "",
		UserName:    "",
		AccountName: accountName,
	}
	return dctx
}

func (g *grpcServer) GetVpnDevice(ctx context.Context, in *infra.GetVpnDeviceIn) (*infra.GetVpnDeviceOut, error) {
	dctx := g.contextFromAccount(ctx, in.AccountName)
	device, err := g.d.GetVPNDevice(dctx, in.ClusterName, in.DeviceName)
	if err != nil {
		return nil, errors.NewE(err)
	}
	wgData, err := json.Marshal(device.WireguardConfig)
	if err != nil {
		return nil, errors.NewE(err)
	}
	wgd, err := json.Marshal(device.Device)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return &infra.GetVpnDeviceOut{
		WgConfig:  wgData,
		VpnDevice: wgd,
	}, nil
}

func (g *grpcServer) UpsertVpnDevice(ctx context.Context, in *infra.UpsertVpnDeviceIn) (*infra.UpsertVpnDeviceOut, error) {
	dctx := g.contextFromAccount(ctx, in.AccountName)
	var wgDevice wireguardV1.Device
	if err := json.Unmarshal(in.VpnDevice, &wgDevice); err != nil {
		return nil, errors.NewE(err)
	}
	device, err := g.d.UpsertManagedVPNDevice(dctx, in.ClusterName, entities.VPNDevice{
		Device:           wgDevice,
		ResourceMetadata: common.ResourceMetadata{},
		AccountName:      in.AccountName,
		ClusterName:      in.ClusterName,
		ManagingByDev:    fn.New(repos.ID(in.Id)),
		SyncStatus:       t.SyncStatus{},
	}, repos.ID(in.Id))

	if err != nil {
		return nil, errors.NewE(err)
	}
	wgData, err := json.Marshal(device.WireguardConfig)
	if err != nil {
		return nil, errors.NewE(err)
	}
	wgd, err := json.Marshal(device.Device)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return &infra.UpsertVpnDeviceOut{
		WgConfig:  wgData,
		VpnDevice: wgd,
	}, nil
}

func (g *grpcServer) DeleteVpnDevice(ctx context.Context, in *infra.DeleteVpnDeviceIn) (*infra.DeleteVpnDeviceOut, error) {
	dctx := g.contextFromAccount(ctx, in.AccountName)
	err := g.d.DeleteManagedVPNDevice(dctx, in.Id)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return &infra.DeleteVpnDeviceOut{}, nil
}

// GetCluster implements infra.InfraServer.
func (g *grpcServer) GetCluster(ctx context.Context, in *infra.GetClusterIn) (*infra.GetClusterOut, error) {
	infraCtx := domain.InfraContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserEmail:   in.UserEmail,
		UserName:    in.UserName,
		AccountName: in.AccountName,
	}
	c, err := g.d.GetCluster(infraCtx, in.ClusterName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if c == nil {
		return nil, errors.Newf("cluster %s not found", in.ClusterName)
	}

	return &infra.GetClusterOut{
		MessageQueueTopic: c.Spec.MessageQueueTopicName,
		DnsHost:           c.Spec.PublicDNSHost,

		IACJobName: func() string {
			if c.Spec.Output != nil {
				return c.Spec.Output.JobName
			}
			return ""
		}(),
		IACJobNamespace: func() string {
			if c.Spec.Output != nil {
				return c.Spec.Output.JobNamespace
			}
			return ""
		}(),
	}, nil
}

// GetNodepool implements infra.InfraServer
func (g *grpcServer) GetNodepool(ctx context.Context, in *infra.GetNodepoolIn) (*infra.GetNodepoolOut, error) {
	infraCtx := domain.InfraContext{
		Context:     ctx,
		UserId:      repos.ID(in.UserId),
		UserEmail:   in.UserEmail,
		UserName:    in.UserName,
		AccountName: in.AccountName,
	}
	np, err := g.d.GetNodePool(infraCtx, in.ClusterName, in.NodepoolName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if np == nil {
		return nil, errors.Newf("nodepool %s not found", in.NodepoolName)
	}

	return &infra.GetNodepoolOut{
		IACJobName:      np.Spec.IAC.JobName,
		IACJobNamespace: np.Spec.IAC.JobNamespace,
	}, nil
}

func newGrpcServer(d domain.Domain) infra.InfraServer {
	return &grpcServer{
		d: d,
	}
}
