package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kloudlite/api/config"
	apiserver "github.com/kloudlite/kloudlite/api/server"
	"github.com/kloudlite/kloudlite/pkg/logger"
	"github.com/spf13/cobra"
)

type serverRunner func(context.Context) error

func newAPIServerCommand() *cobra.Command {
	return newAPIServerCommandWithRunner(runAPIServer)
}

func newAPIServerCommandWithRunner(runner serverRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "Run the API controller and webhook server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner(cmd.Context())
		},
	}
}

func runAPIServer(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
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
