package main

import (
	"fmt"
	"os"

	"github.com/kloudlite/cli/cmd/auth"
	"github.com/urfave/cli/v2"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
	app := &cli.App{
		Name:                 "kl",
		Usage:                "Kloudlite CLI - Manage your Kloudlite platform resources",
		Version:              fmt.Sprintf("%s (commit: %s, built at: %s)", Version, Commit, Date),
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "auth",
				Aliases: []string{"a"},
				Usage:   "Authentication commands",
				Subcommands: []*cli.Command{
					auth.LoginCommand(),
					auth.LogoutCommand(),
					auth.StatusCommand(),
				},
			},
			{
				Name:    "config",
				Aliases: []string{"cfg"},
				Usage:   "Manage CLI configuration",
				Subcommands: []*cli.Command{
					{
						Name:  "show",
						Usage: "Show current configuration",
						Action: func(c *cli.Context) error {
							fmt.Println("Config show functionality not yet implemented")
							return nil
						},
					},
					{
						Name:  "set",
						Usage: "Set configuration value",
						Action: func(c *cli.Context) error {
							fmt.Println("Config set functionality not yet implemented")
							return nil
						},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "Enable debug mode",
				EnvVars: []string{"KLOUDLITE_DEBUG"},
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Config file path",
				Value:   "~/.kloudlite/config.yaml",
				EnvVars: []string{"KLOUDLITE_CONFIG"},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}