package cmd

import (
	"context"
	"fmt"
	"strconv"

	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/spf13/cobra"
)

var exposeCmd = &cobra.Command{
	Use:   "expose <port>",
	Short: "Expose a port from the workspace",
	Long: `Expose a port from the workspace.

The port will be added to the workspace service and an ingress route
will be created with hostname p{port}-{hash}.{subdomain}.
For example: p8080-a1b2c3d4.example.khost.dev`,
	Example: `  # Expose port 8080
  kl expose 8080

  # Expose port 3000
  kl expose 3000

  # List exposed ports
  kl expose list

  # Remove an exposed port
  kl expose remove 3000`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		port, err := parsePort(args[0])
		if err != nil {
			return err
		}
		return handleExpose(port)
	},
}

var exposeListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List exposed ports",
	Long:    `List all ports exposed from the workspace.`,
	Example: `  # List exposed ports
  kl expose list
  kl expose ls`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleExposeList()
	},
}

var exposeRemoveCmd = &cobra.Command{
	Use:     "remove <port>",
	Aliases: []string{"rm"},
	Short:   "Remove an exposed port",
	Long:    `Remove an exposed port from the workspace.`,
	Example: `  # Remove exposed port 3000
  kl expose remove 3000
  kl expose rm 8080`,
	Args: cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getExposedPorts(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := parsePort(args[0])
		if err != nil {
			return err
		}
		return handleExposeRemove(port)
	},
}

func init() {
	// Add subcommands to expose
	exposeCmd.AddCommand(exposeListCmd)
	exposeCmd.AddCommand(exposeRemoveCmd)

	// Register with root command
	RootCmd.AddCommand(exposeCmd)
}

func parsePort(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", portStr)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535, got: %d", port)
	}
	return port, nil
}

func handleExpose(port int) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if port is already exposed
	for _, exposed := range workspace.Spec.Expose {
		if exposed.Port == int32(port) {
			fmt.Printf("Port %d is already exposed\n", port)
			if workspace.Status.Hash != "" && workspace.Status.Subdomain != "" {
				fmt.Printf("URL: https://p%d-%s.%s\n", port, workspace.Status.Hash, workspace.Status.Subdomain)
			}
			return nil
		}
	}

	// Add the new exposed port
	workspace.Spec.Expose = append(workspace.Spec.Expose, workspacesv1.ExposedPort{
		Port: int32(port),
	})

	// Update workspace
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("[✓] Exposing port %d\n", port)

	if workspace.Status.Hash != "" && workspace.Status.Subdomain != "" {
		fmt.Printf("    URL: https://p%d-%s.%s\n", port, workspace.Status.Hash, workspace.Status.Subdomain)
	} else {
		fmt.Printf("    Ingress will be created at p%d-{hash}.{subdomain}\n", port)
	}

	return nil
}

func handleExposeList() error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if len(workspace.Spec.Expose) == 0 {
		fmt.Println("No ports exposed")
		fmt.Println("\nTo expose a port, run:")
		fmt.Println("  kl expose <port>")
		return nil
	}

	fmt.Printf("Exposed ports (%d):\n\n", len(workspace.Spec.Expose))
	for _, exposed := range workspace.Spec.Expose {
		fmt.Printf("  %d", exposed.Port)
		if workspace.Status.Hash != "" && workspace.Status.Subdomain != "" {
			fmt.Printf(" -> https://p%d-%s.%s", exposed.Port, workspace.Status.Hash, workspace.Status.Subdomain)
		}
		fmt.Println()
	}

	return nil
}

func handleExposeRemove(port int) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Find and remove the exposed port
	found := false
	newExpose := []workspacesv1.ExposedPort{}
	for _, exposed := range workspace.Spec.Expose {
		if exposed.Port == int32(port) {
			found = true
			continue
		}
		newExpose = append(newExpose, exposed)
	}

	if !found {
		return fmt.Errorf("port %d is not exposed", port)
	}

	workspace.Spec.Expose = newExpose

	// Update workspace
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("[✓] Removed port %d\n", port)

	return nil
}

// getExposedPorts returns a list of exposed port numbers as strings for shell completion
func getExposedPorts() []string {
	if err := InitClient(); err != nil {
		return nil
	}
	ctx := context.Background()

	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return nil
	}

	var ports []string
	for _, exposed := range workspace.Spec.Expose {
		ports = append(ports, fmt.Sprintf("%d", exposed.Port))
	}
	return ports
}
