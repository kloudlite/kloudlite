package auth

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/kloudlite/cli/pkg/config"
	"github.com/urfave/cli/v2"
)

func LogoutCommand() *cli.Command {
	return &cli.Command{
		Name:  "logout",
		Usage: "Logout from Kloudlite platform",
		Action: func(c *cli.Context) error {
			// Check if logged in
			loggedIn, err := config.IsLoggedIn()
			if err != nil {
				return fmt.Errorf("failed to check login status: %w", err)
			}

			if !loggedIn {
				color.Yellow("You are not logged in")
				return nil
			}

			// Clear config
			if err := config.Clear(); err != nil {
				return fmt.Errorf("failed to logout: %w", err)
			}

			color.Green("✓ Successfully logged out")
			return nil
		},
	}
}