package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud/aws"
)

func main() {
	ctx, cf := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGTERM)
	defer cf()

	p, err := aws.NewProvider(ctx, os.Getenv("K3S_URL"), os.Getenv("K3S_TOKEN"))
	if err != nil {
		slog.Error("create provider failed, ot", "err", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "dry-run":
		{
			slog.Info("calling dry run")
			if err := p.ValidatePermissions(ctx); err != nil {
				slog.Error("failed while calling GetMachineStatus", "err", err)
				os.Exit(1)
			}

			slog.Info("dry run passed")
		}
	case "machine-status":
		{
			slog.Info("calling  machine status")
			mi, err := p.GetMachineStatus(ctx, os.Args[2])
			if err != nil {
				slog.Error("failed while calling GetMachineStatus", "err", err)
				os.Exit(1)
			}

			fmt.Printf("%+v\n", *mi)
		}
	case "stop-machine":
		{
			slog.Info("calling stop machine")
			if err := p.StopMachine(ctx, os.Args[2]); err != nil {
				slog.Error("failed while calling GetMachineStatus", "err", err)
				os.Exit(1)
			}
		}
	case "start-machine":
		{
			slog.Info("calling start machine")
			if err := p.StartMachine(ctx, os.Args[2]); err != nil {
				slog.Error("failed while calling StartMachine", "err", err)
				os.Exit(1)
			}
		}
	case "create-machine":
		{
			slog.Info("calling create machine")
			if err := p.StartMachine(ctx, os.Args[2]); err != nil {
				slog.Error("failed while calling StartMachine", "err", err)
				os.Exit(1)
			}
		}
	}
}
