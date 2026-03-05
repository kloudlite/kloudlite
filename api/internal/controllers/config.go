package controllers

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/kloudlite/kloudlite/api/internal/controllerconfig"
)

// Re-export types from controllerconfig package for backward compatibility
type WorkspaceConfig = controllerconfig.WorkspaceConfig
type EnvironmentConfig = controllerconfig.EnvironmentConfig
type WorkMachineConfig = controllerconfig.WorkMachineConfig
type WMIngressConfig = controllerconfig.WMIngressConfig
type SnapshotConfig = controllerconfig.SnapshotConfig
type ControllerConfig = controllerconfig.ControllerConfig

// LoadConfig loads controller configuration from environment variables
func LoadConfig() (*ControllerConfig, error) {
	var cfg ControllerConfig
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse controller configuration: %w", err)
	}

	// Set defaults for duration fields if not set via environment
	if cfg.Environment.PodTerminationRetryInterval == 0 {
		cfg.Environment.PodTerminationRetryInterval = 2 * time.Second
	}
	if cfg.Environment.SnapshotRestoreRetryInterval == 0 {
		cfg.Environment.SnapshotRestoreRetryInterval = 2 * time.Second
	}
	if cfg.Environment.SnapshotRequestRetryInterval == 0 {
		cfg.Environment.SnapshotRequestRetryInterval = 2 * time.Second
	}
	if cfg.Environment.ForkRetryInterval == 0 {
		cfg.Environment.ForkRetryInterval = 5 * time.Second
	}
	if cfg.Environment.StatusUpdateRetryInterval == 0 {
		cfg.Environment.StatusUpdateRetryInterval = 5 * time.Second
	}
	if cfg.Environment.DeletionRetryInterval == 0 {
		cfg.Environment.DeletionRetryInterval = 5 * time.Second
	}
	if cfg.Environment.LifecycleRetryInterval == 0 {
		cfg.Environment.LifecycleRetryInterval = 5 * time.Second
	}

	if cfg.WorkMachine.CloudOperationRetryInterval == 0 {
		cfg.WorkMachine.CloudOperationRetryInterval = 5 * time.Second
	}
	if cfg.WorkMachine.MachineStatusCheckInterval == 0 {
		cfg.WorkMachine.MachineStatusCheckInterval = 5 * time.Second
	}
	if cfg.WorkMachine.MachineStartupRetryInterval == 0 {
		cfg.WorkMachine.MachineStartupRetryInterval = 10 * time.Second
	}
	if cfg.WorkMachine.NodeJoinRetryInterval == 0 {
		cfg.WorkMachine.NodeJoinRetryInterval = 10 * time.Second
	}
	if cfg.WorkMachine.VolumeResizeRetryInterval == 0 {
		cfg.WorkMachine.VolumeResizeRetryInterval = 10 * time.Second
	}
	if cfg.WorkMachine.MachineTypeChangeRetryInterval == 0 {
		cfg.WorkMachine.MachineTypeChangeRetryInterval = 5 * time.Second
	}
	if cfg.WorkMachine.AutoShutdownCheckInterval == 0 {
		cfg.WorkMachine.AutoShutdownCheckInterval = 5 * time.Minute
	}
	if cfg.WorkMachine.CloudMachineCreationRetryInterval == 0 {
		cfg.WorkMachine.CloudMachineCreationRetryInterval = 2 * time.Second
	}
	if cfg.WorkMachine.CloudMachineStopRetryInterval == 0 {
		cfg.WorkMachine.CloudMachineStopRetryInterval = 10 * time.Second
	}
	if cfg.WorkMachine.CloudMachineStartRetryInterval == 0 {
		cfg.WorkMachine.CloudMachineStartRetryInterval = 10 * time.Second
	}
	if cfg.WorkMachine.NodeDrainRetryInterval == 0 {
		cfg.WorkMachine.NodeDrainRetryInterval = 5 * time.Second
	}
	if cfg.WorkMachine.NodeDeleteRetryInterval == 0 {
		cfg.WorkMachine.NodeDeleteRetryInterval = 2 * time.Second
	}
	if cfg.WorkMachine.VolumeResizeCheckInterval == 0 {
		cfg.WorkMachine.VolumeResizeCheckInterval = 10 * time.Second
	}
	if cfg.WorkMachine.NodeJoinCheckInterval == 0 {
		cfg.WorkMachine.NodeJoinCheckInterval = 10 * time.Second
	}
	if cfg.WorkMachine.NodeReadyRetryInterval == 0 {
		cfg.WorkMachine.NodeReadyRetryInterval = 5 * time.Second
	}
	if cfg.WorkMachine.AutoShutdownTriggerRetryInterval == 0 {
		cfg.WorkMachine.AutoShutdownTriggerRetryInterval = 5 * time.Second
	}

	if cfg.WMIngress.ProxyTimeout == 0 {
		cfg.WMIngress.ProxyTimeout = 30 * time.Second
	}
	if cfg.WMIngress.ProxyKeepAlive == 0 {
		cfg.WMIngress.ProxyKeepAlive = 30 * time.Second
	}
	if cfg.WMIngress.ProxyIdleConnTimeout == 0 {
		cfg.WMIngress.ProxyIdleConnTimeout = 90 * time.Second
	}
	if cfg.WMIngress.ProxyTLSHandshakeTimeout == 0 {
		cfg.WMIngress.ProxyTLSHandshakeTimeout = 10 * time.Second
	}
	if cfg.WMIngress.ProxyExpectContinueTimeout == 0 {
		cfg.WMIngress.ProxyExpectContinueTimeout = 1 * time.Second
	}

	if cfg.Snapshot.DefaultRequeueInterval == 0 {
		cfg.Snapshot.DefaultRequeueInterval = 30 * time.Second
	}
	if cfg.Snapshot.StorageCleanupRetryInterval == 0 {
		cfg.Snapshot.StorageCleanupRetryInterval = 5 * time.Second
	}
	if cfg.Snapshot.SnapshotReadyCheckInterval == 0 {
		cfg.Snapshot.SnapshotReadyCheckInterval = 5 * time.Second
	}
	if cfg.Snapshot.SnapshotRestoreWaitInterval == 0 {
		cfg.Snapshot.SnapshotRestoreWaitInterval = 5 * time.Second
	}
	if cfg.Snapshot.SnapshotRestoreStatusRetryInterval == 0 {
		cfg.Snapshot.SnapshotRestoreStatusRetryInterval = 5 * time.Second
	}

	// Set derived fields for Workspace config
	cfg.Workspace.DefaultRequeueInterval = time.Duration(cfg.Workspace.RequeueIntervalMinutes) * time.Minute
	cfg.Workspace.RBACCleanupRetryInterval = time.Duration(cfg.Workspace.RBACCleanupIntervalMinutes) * time.Minute

	// Set derived fields for Environment config
	cfg.Environment.DefaultRequeueInterval = 5 * time.Second
	cfg.Environment.StatefulSetScaleTimeout = 5 * time.Minute
	cfg.Environment.SnapshotRestoreWaitInterval = cfg.Environment.SnapshotRestoreRetryInterval

	return &cfg, nil
}
