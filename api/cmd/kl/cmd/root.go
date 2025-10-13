package cmd

import (
	"fmt"

	"github.com/kloudlite/kloudlite/api/cmd/kl/pkg/workspace"
	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var (
	// Global workspace client
	WsClient *workspace.Client
)

// RootCmd represents the base command
var RootCmd = &cobra.Command{
	Use:   "kl",
	Short: "Kloudlite Workspace Manager",
	Long: `kl is a CLI tool for managing Kloudlite workspaces.

It provides commands to manage workspace configuration, packages, and status
directly from within your workspace container.`,
	Example: `  # Show workspace status
  kl status
  kl st

  # Add packages
  kl pkg add git vim
  kl p a nodejs

  # List installed packages
  kl pkg list
  kl p ls`,
}

func init() {
	// Add commands
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(pkgCmd)
	RootCmd.AddCommand(configCmd)
}

// InitClient initializes the workspace client if not already initialized
func InitClient() error {
	if WsClient != nil {
		return nil
	}

	var err error
	WsClient, err = workspace.New()
	if err != nil {
		return fmt.Errorf("failed to initialize workspace client: %w", err)
	}

	return nil
}

// Execute runs the root command
func Execute() error {
	return RootCmd.Execute()
}
