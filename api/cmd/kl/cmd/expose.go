package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var httpExposeCmd = &cobra.Command{
	Use:   "http-expose <port>",
	Short: "Expose an HTTP port from the workspace",
	Long: `Expose an HTTP port from the workspace via ingress.

An ingress route will be created with hostname p{port}-{hash}.{subdomain}.
For example: p3000-a1b2c3d4.example.khost.dev

Note: TCP and UDP ports are already accessible via the headless service.
This command is only needed to create public HTTP ingress routes.`,
	Example: `  # Expose HTTP port 3000
  kl http-expose 3000

  # Expose HTTP port 8080
  kl http-expose 8080

  # List exposed HTTP ports
  kl http-expose list

  # Remove an exposed HTTP port
  kl http-expose remove 3000`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		port, err := parsePort(args[0])
		if err != nil {
			return err
		}
		return handleHttpExpose(port)
	},
}

var httpExposeListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List exposed HTTP ports",
	Long:    `List all HTTP ports exposed from the workspace via ingress.`,
	Example: `  # List exposed HTTP ports
  kl http-expose list
  kl http-expose ls`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleHttpExposeList()
	},
}

var httpExposeRemoveCmd = &cobra.Command{
	Use:     "remove <port>",
	Aliases: []string{"rm"},
	Short:   "Remove an exposed HTTP port",
	Long:    `Remove an HTTP port exposure from the workspace.`,
	Example: `  # Remove exposed HTTP port 3000
  kl http-expose remove 3000
  kl http-expose rm 8080`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := parsePort(args[0])
		if err != nil {
			return err
		}
		return handleHttpExposeRemove(port)
	},
}

func init() {
	// Add subcommands to http-expose
	httpExposeCmd.AddCommand(httpExposeListCmd)
	httpExposeCmd.AddCommand(httpExposeRemoveCmd)

	// Register with root command
	RootCmd.AddCommand(httpExposeCmd)
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

func handleHttpExpose(port int) error {
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
	for _, exposedPort := range workspace.Spec.HttpExpose {
		if exposedPort == int32(port) {
			fmt.Printf("Port %d is already exposed\n", port)
			if workspace.Status.Hash != "" && workspace.Status.Subdomain != "" {
				fmt.Printf("URL: https://p%d-%s.%s\n", port, workspace.Status.Hash, workspace.Status.Subdomain)
			}
			return nil
		}
	}

	// Add the new exposed port
	workspace.Spec.HttpExpose = append(workspace.Spec.HttpExpose, int32(port))

	// Update workspace
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("[✓] Exposing HTTP port %d\n", port)

	if workspace.Status.Hash != "" && workspace.Status.Subdomain != "" {
		fmt.Printf("    URL: https://p%d-%s.%s\n", port, workspace.Status.Hash, workspace.Status.Subdomain)
	} else {
		fmt.Printf("    Ingress will be created at p%d-{hash}.{subdomain}\n", port)
	}

	return nil
}

func handleHttpExposeList() error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	if len(workspace.Spec.HttpExpose) == 0 {
		fmt.Println("No HTTP ports exposed")
		fmt.Println("\nTo expose an HTTP port, run:")
		fmt.Println("  kl http-expose <port>")
		return nil
	}

	fmt.Printf("Exposed HTTP ports (%d):\n\n", len(workspace.Spec.HttpExpose))
	for _, port := range workspace.Spec.HttpExpose {
		fmt.Printf("  %d", port)
		if workspace.Status.Hash != "" && workspace.Status.Subdomain != "" {
			fmt.Printf(" -> https://p%d-%s.%s", port, workspace.Status.Hash, workspace.Status.Subdomain)
		}
		fmt.Println()
	}

	return nil
}

func handleHttpExposeRemove(port int) error {
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
	newExpose := []int32{}
	for _, exposedPort := range workspace.Spec.HttpExpose {
		if exposedPort == int32(port) {
			found = true
			continue
		}
		newExpose = append(newExpose, exposedPort)
	}

	if !found {
		return fmt.Errorf("port %d is not exposed", port)
	}

	workspace.Spec.HttpExpose = newExpose

	// Update workspace
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("[✓] Removed HTTP port %d\n", port)

	return nil
}
