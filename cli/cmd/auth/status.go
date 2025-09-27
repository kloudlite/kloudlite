package auth

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/kloudlite/cli/pkg/config"
	"github.com/urfave/cli/v2"
)

func StatusCommand() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Check authentication status",
		Action: func(c *cli.Context) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if !cfg.IsAuthenticated() {
				color.Yellow("You are not logged in")
				fmt.Println("\nTo login, run:")
				color.Cyan("  kl auth login")
				return nil
			}

			color.Green("✓ Logged in")
			fmt.Println()
			if cfg.UserID != "" {
				fmt.Printf("User ID: %s\n", cfg.UserID)
			}
			if cfg.ServerAddr != "" {
				fmt.Printf("Server: %s\n", cfg.ServerAddr)
			}

			return nil
		},
	}
}