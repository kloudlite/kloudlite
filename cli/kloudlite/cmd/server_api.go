package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kloudlite/api/config"
	"github.com/kloudlite/kloudlite/api/crds"
	"github.com/kloudlite/kloudlite/api/k8s"
	apiserver "github.com/kloudlite/kloudlite/api/server"
	"github.com/kloudlite/kloudlite/pkg/logger"
	"github.com/spf13/cobra"
)

type serverRunner func(context.Context) error

type apiServerOptions struct {
	InstallCRDs bool
	CRDsDir     string
}

func newAPIServerCommand() *cobra.Command {
	return newAPIServerCommandWithRunner(runAPIServer)
}

func newAPIServerCommandWithRunner(runner serverRunner) *cobra.Command {
	opts := &apiServerOptions{}
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Run the API controller and webhook server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.InstallCRDs {
				return runAPIServerWithOptions(cmd.Context(), opts)
			}
			return runner(cmd.Context())
		},
	}
	cmd.Flags().BoolVar(&opts.InstallCRDs, "install-crds", false, "Install Kloudlite CRDs before starting the API server")
	cmd.Flags().StringVar(&opts.CRDsDir, "crds-dir", "", "Optional directory containing Kloudlite CRD YAML manifests; defaults to embedded CRD assets")
	return cmd
}

func runAPIServer(ctx context.Context) error {
	return runAPIServerWithOptions(ctx, &apiServerOptions{})
}

func runAPIServerWithOptions(ctx context.Context, opts *apiServerOptions) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if opts != nil && opts.InstallCRDs {
		if err := crds.Install(ctx, &k8s.ClientOptions{KubeconfigPath: cfg.Kubernetes.KubeconfigPath, Context: cfg.Kubernetes.Context, MasterURL: cfg.Kubernetes.MasterURL}, opts.CRDsDir); err != nil {
			return err
		}
	}

	appLogger, err := logger.New(cfg.LogLevel, cfg.Environment)
	if err != nil {
		return err
	}
	defer appLogger.Sync()

	srv := apiserver.New(cfg, appLogger)

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case <-ctx.Done():
		return srv.Shutdown(context.Background())
	case <-quit:
		return srv.Shutdown(context.Background())
	case err := <-serverErr:
		return err
	}
}
