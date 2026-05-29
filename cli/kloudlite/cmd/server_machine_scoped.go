package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kloudlite/api/config"
	"go.uber.org/zap"
	"github.com/kloudlite/kloudlite/api/k8s"
	"github.com/kloudlite/kloudlite/api/tunnel"
	workmachinenodemanager "github.com/kloudlite/kloudlite/cli/workmachine-node-manager"
	"github.com/kloudlite/kloudlite/controllers"
	"github.com/kloudlite/kloudlite/pkg/logger"
	"github.com/spf13/cobra"
)

func newWorkMachineManagerCommand() *cobra.Command {
	return newWorkMachineManagerCommandWithRunner(runWorkMachineManager)
}

func newWorkMachineManagerCommandWithRunner(runner serverRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "workmachine-manager",
		Short: "Run WorkMachine manager controllers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner(cmd.Context())
		},
	}
}

func runWorkMachineManager(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	k8sClient, err := k8s.NewClient(ctx, &k8s.ClientOptions{KubeconfigPath: cfg.Kubernetes.KubeconfigPath, Context: cfg.Kubernetes.Context, MasterURL: cfg.Kubernetes.MasterURL})
	if err != nil {
		return err
	}

	appLogger, err := logger.New(cfg.LogLevel, cfg.Environment)
	if err != nil {
		return err
	}
	defer appLogger.Sync()

	mgr, err := controllers.NewMachineScopedManager(k8sClient.Config, &cfg.Installation, &cfg.Auth, appLogger)
	if err != nil {
		return err
	}
	if err := mgr.StartIngressServers(ctx); err != nil {
		return err
	}

	managerErr := make(chan error, 1)
	go func() {
		managerErr <- mgr.Start(ctx)
	}()
	go func() {
		managerErr <- workmachinenodemanager.Start(ctx, workmachinenodemanager.Config{
			Namespace:                os.Getenv("NAMESPACE"),
			WorkMachineName:          os.Getenv("WORKMACHINE_NAME"),
			SnapshotRegistryEndpoint: os.Getenv("SNAPSHOT_REGISTRY_ENDPOINT"),
			SnapshotRegistryPrefix:   os.Getenv("SNAPSHOT_REGISTRY_PREFIX"),
			SnapshotRegistryInsecure: os.Getenv("SNAPSHOT_REGISTRY_INSECURE"),
		})
	}()
	go func() {
		tunnelArgs := []string{
			"--listen", ":443",
			"--tls-secret", "kloudlite-wildcard-cert-tls",
			"--ca-cert-secret", "kloudlite-wildcard-cert-tls",
			"--kltun-tls-secret", "kloudlite-wildcard-cert-tls",
			"--wireguard-target", "127.0.0.1:51820",
			"--watch-config",
			"--config-path", "/etc/wireguard/wg0.conf",
			"--namespace", os.Getenv("NAMESPACE"),
			"--router-service", "wm-ingress-controller",
			"--dns-listen", ":53",
			"--upstream-dns", "10.43.0.10:53",
		}
		appLogger.Info("starting tunnel-server (inline)", zap.Any("config", tunnelArgs))
		if err := tunnel.Run(ctx, tunnelArgs); err != nil {
			appLogger.Error("failed to start tunnel-server", zap.Error(err))
			managerErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case <-ctx.Done():
		return nil
	case <-quit:
		return nil
	case err := <-managerErr:
		return err
	}
}
