package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kloudlite/api/config"
	"github.com/kloudlite/kloudlite/api/k8s"
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

	managerErr := make(chan error, 1)
	go func() {
		managerErr <- mgr.Start(ctx)
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
