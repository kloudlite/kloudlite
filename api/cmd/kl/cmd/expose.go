package cmd

import (
	"context"
	"fmt"
	"strconv"

	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/spf13/cobra"
)

var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Expose ports from the workspace",
	Long: `Expose ports from the workspace using different protocols.

TCP/UDP ports are added to the workspace service for direct access.
HTTP ports also get an ingress route with hostname p{port}-{hash}.{subdomain}.`,
	Example: `  # Expose a TCP port
  kl expose tcp 3000

  # Expose a UDP port
  kl expose udp 5353

  # Expose an HTTP port (creates ingress at p8080-{hash}.{subdomain})
  kl expose http 8080

  # List exposed ports
  kl expose list

  # Remove an exposed port
  kl expose remove tcp 3000`,
}

var exposeTCPCmd = &cobra.Command{
	Use:   "tcp <port>",
	Short: "Expose a TCP port",
	Long:  `Expose a TCP port from the workspace. The port will be added to the workspace service.`,
	Example: `  # Expose TCP port 3000
  kl expose tcp 3000`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := parsePort(args[0])
		if err != nil {
			return err
		}
		return handleExpose(workspacesv1.ExposeProtocolTCP, port)
	},
}

var exposeUDPCmd = &cobra.Command{
	Use:   "udp <port>",
	Short: "Expose a UDP port",
	Long:  `Expose a UDP port from the workspace. The port will be added to the workspace service.`,
	Example: `  # Expose UDP port 5353
  kl expose udp 5353`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := parsePort(args[0])
		if err != nil {
			return err
		}
		return handleExpose(workspacesv1.ExposeProtocolUDP, port)
	},
}

var exposeHTTPCmd = &cobra.Command{
	Use:   "http <port>",
	Short: "Expose an HTTP port",
	Long: `Expose an HTTP port from the workspace.

The port will be added to the workspace service and an ingress route
will be created with hostname p{port}-{hash}.{subdomain}.
For example: p8080-a1b2c3d4.example.khost.dev`,
	Example: `  # Expose HTTP port 8080
  kl expose http 8080`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		port, err := parsePort(args[0])
		if err != nil {
			return err
		}
		return handleExpose(workspacesv1.ExposeProtocolHTTP, port)
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
	Use:     "remove <protocol> <port>",
	Aliases: []string{"rm"},
	Short:   "Remove an exposed port",
	Long:    `Remove an exposed port from the workspace.`,
	Example: `  # Remove exposed TCP port 3000
  kl expose remove tcp 3000
  kl expose rm http 8080`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		protocol := args[0]
		port, err := parsePort(args[1])
		if err != nil {
			return err
		}
		var exposeProtocol workspacesv1.ExposeProtocol
		switch protocol {
		case "tcp":
			exposeProtocol = workspacesv1.ExposeProtocolTCP
		case "udp":
			exposeProtocol = workspacesv1.ExposeProtocolUDP
		case "http":
			exposeProtocol = workspacesv1.ExposeProtocolHTTP
		default:
			return fmt.Errorf("invalid protocol: %s (must be tcp, udp, or http)", protocol)
		}
		return handleExposeRemove(exposeProtocol, port)
	},
}

func init() {
	// Add subcommands to expose
	exposeCmd.AddCommand(exposeTCPCmd)
	exposeCmd.AddCommand(exposeUDPCmd)
	exposeCmd.AddCommand(exposeHTTPCmd)
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

func handleExpose(protocol workspacesv1.ExposeProtocol, port int) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if port is already exposed with same protocol
	for _, exposed := range workspace.Spec.Expose {
		if exposed.Port == int32(port) && exposed.Protocol == protocol {
			fmt.Printf("Port %d/%s is already exposed\n", port, protocol)
			if protocol == workspacesv1.ExposeProtocolHTTP && workspace.Status.Hash != "" && workspace.Status.Subdomain != "" {
				fmt.Printf("URL: https://p%d-%s.%s\n", port, workspace.Status.Hash, workspace.Status.Subdomain)
			}
			return nil
		}
	}

	// Add the new exposed port
	workspace.Spec.Expose = append(workspace.Spec.Expose, workspacesv1.ExposedPort{
		Port:     int32(port),
		Protocol: protocol,
	})

	// Update workspace
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("[✓] Exposing %s port %d\n", protocol, port)

	if protocol == workspacesv1.ExposeProtocolHTTP {
		if workspace.Status.Hash != "" && workspace.Status.Subdomain != "" {
			fmt.Printf("    URL: https://p%d-%s.%s\n", port, workspace.Status.Hash, workspace.Status.Subdomain)
		} else {
			fmt.Printf("    Ingress will be created at p%d-{hash}.{subdomain}\n", port)
		}
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
		fmt.Println("  kl expose tcp <port>")
		fmt.Println("  kl expose udp <port>")
		fmt.Println("  kl expose http <port>")
		return nil
	}

	fmt.Printf("Exposed ports (%d):\n\n", len(workspace.Spec.Expose))
	for _, exposed := range workspace.Spec.Expose {
		fmt.Printf("  %s/%d", exposed.Protocol, exposed.Port)
		if exposed.Protocol == workspacesv1.ExposeProtocolHTTP {
			if workspace.Status.Hash != "" && workspace.Status.Subdomain != "" {
				fmt.Printf(" -> https://p%d-%s.%s", exposed.Port, workspace.Status.Hash, workspace.Status.Subdomain)
			}
		}
		fmt.Println()
	}

	return nil
}

func handleExposeRemove(protocol workspacesv1.ExposeProtocol, port int) error {
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
		if exposed.Port == int32(port) && exposed.Protocol == protocol {
			found = true
			continue
		}
		newExpose = append(newExpose, exposed)
	}

	if !found {
		return fmt.Errorf("port %d/%s is not exposed", port, protocol)
	}

	workspace.Spec.Expose = newExpose

	// Update workspace
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("[✓] Removed %s port %d\n", protocol, port)

	return nil
}
