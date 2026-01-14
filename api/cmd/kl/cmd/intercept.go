package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	fzf "github.com/junegunn/fzf/src"
	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var interceptCmd = &cobra.Command{
	Use:     "intercept",
	Aliases: []string{"i", "int"},
	Short:   "Manage service interception",
	Long: `Manage service interception to redirect service traffic to your workspace.

Service interception allows you to redirect traffic from a service in an environment
to your workspace for local development and testing.`,
	Example: `  # Start intercepting a service (interactive)
  kl intercept start
  kl i s

  # Start intercepting a specific service
  kl intercept start api-server
  kl i s api-server

  # List active intercepts
  kl intercept list
  kl i ls

  # Stop intercepting a service
  kl intercept stop api-server
  kl i sp api-server

  # Show intercept status
  kl intercept status
  kl i st`,
}

var interceptStartCmd = &cobra.Command{
	Use:     "start [service-name]",
	Aliases: []string{"s", "begin"},
	Short:   "Start intercepting a service",
	Long: `Start intercepting a service to redirect its traffic to your workspace.

If no service name is provided, an interactive list will be shown.
You will be prompted to map each service port to a workspace port.`,
	Example: `  # Interactive service selection
  kl intercept start
  kl i s

  # Intercept specific service
  kl intercept start api-server
  kl i s api-server`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getAvailableServiceNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Interactive mode
			return handleInterceptStartInteractive()
		}
		// Direct mode
		return handleInterceptStart(args[0])
	},
}

var interceptStopCmd = &cobra.Command{
	Use:     "stop [service-name]",
	Aliases: []string{"sp", "end"},
	Short:   "Stop intercepting a service",
	Long: `Stop intercepting a service and restore normal traffic routing.

If no service name is provided, an interactive list of active intercepts will be shown.`,
	Example: `  # Interactive intercept selection
  kl intercept stop
  kl i sp

  # Stop specific service intercept
  kl intercept stop api-server
  kl i sp api-server`,
	Args: cobra.MaximumNArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getActiveInterceptNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Interactive mode
			return handleInterceptStopInteractive()
		}
		return handleInterceptStop(args[0])
	},
}

var interceptListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List active service intercepts",
	Long:    `List all active service intercepts in the connected environment.`,
	Example: `  # List intercepts
  kl intercept list
  kl i ls`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleInterceptList()
	},
}

var interceptStatusCmd = &cobra.Command{
	Use:     "status [service-name]",
	Aliases: []string{"st"},
	Short:   "Show intercept status",
	Long:    `Show detailed status of service intercept(s).`,
	Example: `  # Show all intercepts status
  kl intercept status
  kl i st

  # Show specific service intercept status
  kl intercept status api-server
  kl i st api-server`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return handleInterceptStatus("")
		}
		return handleInterceptStatus(args[0])
	},
}

func init() {
	// Add subcommands
	interceptCmd.AddCommand(interceptStartCmd)
	interceptCmd.AddCommand(interceptStopCmd)
	interceptCmd.AddCommand(interceptListCmd)
	interceptCmd.AddCommand(interceptStatusCmd)

	// Register with root command
	RootCmd.AddCommand(interceptCmd)
}

// EnvironmentService represents a service from an environment's compose
type EnvironmentService struct {
	EnvironmentName string
	ServiceName     string
	Ports           []int32
	State           string
}

// getConnectedEnvironment gets the environment that the workspace is connected to
func getConnectedEnvironment(ctx context.Context, envName, workspaceNamespace string) (*environmentv1.Environment, error) {
	env := &environmentv1.Environment{}
	if err := WsClient.K8sClient.Get(ctx, client.ObjectKey{
		Name:      envName,
		Namespace: workspaceNamespace,
	}, env); err != nil {
		return nil, err
	}
	return env, nil
}

// listEnvironmentServices lists all services from the connected environment's compose
func listEnvironmentServices(ctx context.Context, env *environmentv1.Environment) ([]EnvironmentService, error) {
	if env.Status.ComposeStatus == nil {
		return nil, nil
	}

	var services []EnvironmentService
	for _, svcStatus := range env.Status.ComposeStatus.Services {
		services = append(services, EnvironmentService{
			EnvironmentName: env.Name,
			ServiceName:     svcStatus.Name,
			Ports:           svcStatus.Ports,
			State:           svcStatus.State,
		})
	}

	return services, nil
}

// findServiceByName finds a service by name in the connected environment
func findServiceByName(ctx context.Context, serviceName string, env *environmentv1.Environment) (*EnvironmentService, error) {
	services, err := listEnvironmentServices(ctx, env)
	if err != nil {
		return nil, err
	}

	for _, svc := range services {
		if svc.ServiceName == serviceName {
			return &svc, nil
		}
	}

	return nil, fmt.Errorf("service '%s' not found in environment", serviceName)
}

// findActiveInterceptByServiceName finds an active intercept by service name
func findActiveInterceptByServiceName(ctx context.Context, serviceName string, env *environmentv1.Environment, workspaceName string) (*ActiveIntercept, error) {
	intercepts := listActiveInterceptsFromEnv(env, workspaceName)

	for _, intercept := range intercepts {
		if intercept.ServiceName == serviceName {
			return &intercept, nil
		}
	}

	return nil, fmt.Errorf("no active intercept found for service '%s'", serviceName)
}

func handleInterceptStartInteractive() error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is connected to an environment
	if workspace.Status.ConnectedEnvironment == nil {
		return fmt.Errorf("workspace is not connected to any environment. Connect using 'kl env connect' first")
	}

	envName := workspace.Status.ConnectedEnvironment.Name
	if envName == "" {
		return fmt.Errorf("workspace has no connected environment")
	}

	// Get the connected environment
	env, err := getConnectedEnvironment(ctx, envName, workspace.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get environment '%s': %w", envName, err)
	}

	// List services from environment's compose
	fmt.Println("Loading services...")
	services, err := listEnvironmentServices(ctx, env)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		return fmt.Errorf("no services found in environment")
	}

	// Use embedded fzf to select service
	selectedService, err := selectEnvironmentServiceWithFzf(services)
	if err != nil {
		return err
	}

	return handleInterceptStartWithService(ctx, env, *selectedService, workspace.Name, workspace.Namespace)
}

func handleInterceptStart(serviceName string) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is connected to an environment
	if workspace.Status.ConnectedEnvironment == nil {
		return fmt.Errorf("workspace is not connected to any environment. Connect using 'kl env connect' first")
	}

	envName := workspace.Status.ConnectedEnvironment.Name
	if envName == "" {
		return fmt.Errorf("workspace has no connected environment")
	}

	// Get the connected environment
	env, err := getConnectedEnvironment(ctx, envName, workspace.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get environment '%s': %w", envName, err)
	}

	// Find service by name in environment
	svc, err := findServiceByName(ctx, serviceName, env)
	if err != nil {
		return err
	}

	return handleInterceptStartWithService(ctx, env, *svc, workspace.Name, workspace.Namespace)
}

func handleInterceptStartWithService(ctx context.Context, env *environmentv1.Environment, svc EnvironmentService, workspaceName, workspaceNamespace string) error {
	// Check if compose exists
	if env.Spec.Compose == nil {
		return fmt.Errorf("environment has no compose configuration")
	}

	// Check if intercept already exists in compose spec
	for _, intercept := range env.Spec.Compose.Intercepts {
		if intercept.ServiceName == svc.ServiceName && intercept.Enabled {
			if intercept.WorkspaceRef != nil {
				return fmt.Errorf("service '%s' is already being intercepted by workspace '%s'", svc.ServiceName, intercept.WorkspaceRef.Name)
			}
			return fmt.Errorf("service '%s' is already being intercepted", svc.ServiceName)
		}
	}

	// Prompt for port mappings
	var portMappings []environmentv1.PortMapping
	reader := bufio.NewReader(os.Stdin)

	if len(svc.Ports) > 0 {
		// Service has defined ports - let user select which to intercept
		selectedPorts, err := selectPortsWithFzf(svc.ServiceName, svc.Ports)
		if err != nil {
			return err
		}

		if len(selectedPorts) == 0 {
			return fmt.Errorf("no ports selected for interception")
		}

		portMappings = make([]environmentv1.PortMapping, 0, len(selectedPorts))

		// Ask for workspace port mapping for each selected port
		fmt.Printf("\nMap selected ports to workspace ports:\n")
		for _, port := range selectedPorts {
			fmt.Printf("  Service port %d → Workspace port [%d]: ", port, port)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			input = strings.TrimSpace(input)
			workspacePort := port
			if input != "" {
				parsedPort, err := strconv.Atoi(input)
				if err != nil {
					return fmt.Errorf("invalid port number: %s", input)
				}
				workspacePort = int32(parsedPort)
			}

			portMappings = append(portMappings, environmentv1.PortMapping{
				ServicePort:   port,
				WorkspacePort: workspacePort,
				Protocol:      corev1.ProtocolTCP,
			})
		}
	} else {
		// No ports defined - ask user to manually configure
		fmt.Printf("\nService '%s' has no ports defined.\n", svc.ServiceName)
		fmt.Println("You need to manually specify port mappings.")
		fmt.Println("Enter port mappings (press Enter with empty service port to finish):")

		portMappings = make([]environmentv1.PortMapping, 0)

		for {
			fmt.Print("  Service port (or press Enter to finish): ")
			servicePortInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			servicePortInput = strings.TrimSpace(servicePortInput)
			if servicePortInput == "" {
				if len(portMappings) == 0 {
					return fmt.Errorf("at least one port mapping is required")
				}
				break
			}

			servicePort, err := strconv.Atoi(servicePortInput)
			if err != nil || servicePort < 1 || servicePort > 65535 {
				fmt.Println("  Invalid port number. Please enter a number between 1 and 65535.")
				continue
			}

			fmt.Print("  Workspace port: ")
			workspacePortInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			workspacePortInput = strings.TrimSpace(workspacePortInput)
			if workspacePortInput == "" {
				fmt.Println("  Workspace port is required.")
				continue
			}

			workspacePort, err := strconv.Atoi(workspacePortInput)
			if err != nil || workspacePort < 1 || workspacePort > 65535 {
				fmt.Println("  Invalid port number. Please enter a number between 1 and 65535.")
				continue
			}

			portMappings = append(portMappings, environmentv1.PortMapping{
				ServicePort:   int32(servicePort),
				WorkspacePort: int32(workspacePort),
				Protocol:      corev1.ProtocolTCP,
			})

			fmt.Printf("  [✓] Added mapping: %d (service) → %d (workspace)\n\n", servicePort, workspacePort)
		}
	}

	// Add intercept to composition spec
	interceptConfig := environmentv1.ServiceInterceptConfig{
		ServiceName:  svc.ServiceName,
		PortMappings: portMappings,
		Enabled:      true,
		WorkspaceRef: &corev1.ObjectReference{
			Name:      workspaceName,
			Namespace: workspaceNamespace,
		},
	}

	// Check if intercept config already exists (but was disabled) and update it
	found := false
	for i, existing := range env.Spec.Compose.Intercepts {
		if existing.ServiceName == svc.ServiceName {
			env.Spec.Compose.Intercepts[i] = interceptConfig
			found = true
			break
		}
	}
	if !found {
		env.Spec.Compose.Intercepts = append(env.Spec.Compose.Intercepts, interceptConfig)
	}

	// Update environment
	if err := WsClient.K8sClient.Update(ctx, env); err != nil {
		return fmt.Errorf("failed to update environment: %w", err)
	}

	fmt.Printf("\n[✓] Service intercept added\n")

	// Wait for the intercept to become active
	if err := waitForInterceptSync(ctx, env.Name, svc.ServiceName, workspaceNamespace, "start"); err != nil {
		return fmt.Errorf("service intercept activation failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("[✓] Service intercept is now active\n")
	fmt.Printf("Service '%s' is being intercepted\n\n", svc.ServiceName)
	fmt.Println("Port mappings:")
	for _, mapping := range portMappings {
		fmt.Printf("  %d (service) → %d (workspace)\n", mapping.ServicePort, mapping.WorkspacePort)
	}
	fmt.Printf("\nTraffic to '%s' is now routed to your workspace.\n", svc.ServiceName)

	return nil
}

func handleInterceptStopInteractive() error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is connected to an environment
	if workspace.Status.ConnectedEnvironment == nil {
		return fmt.Errorf("workspace is not connected to any environment. Connect using 'kl env connect' first")
	}

	envName := workspace.Status.ConnectedEnvironment.Name
	if envName == "" {
		return fmt.Errorf("workspace has no connected environment")
	}

	// Get the connected environment
	env, err := getConnectedEnvironment(ctx, envName, workspace.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get environment '%s': %w", envName, err)
	}

	// Get all active intercepts from environment
	activeIntercepts := listActiveInterceptsFromEnv(env, workspace.Name)

	if len(activeIntercepts) == 0 {
		return fmt.Errorf("no active service intercepts found")
	}

	// Use embedded fzf to select intercept
	selectedIntercept, err := selectActiveInterceptWithFzf(activeIntercepts)
	if err != nil {
		return err
	}

	return handleInterceptStopWithEnv(ctx, env, selectedIntercept.ServiceName, workspace.Namespace)
}

// ActiveIntercept represents an active intercept for display
type ActiveIntercept struct {
	EnvironmentName string
	ServiceName     string
	Phase           string
	Message         string
}

// listActiveInterceptsFromEnv lists all active intercepts from an environment's compose status
func listActiveInterceptsFromEnv(env *environmentv1.Environment, workspaceName string) []ActiveIntercept {
	if env.Status.ComposeStatus == nil {
		return nil
	}

	var intercepts []ActiveIntercept
	for _, intercept := range env.Status.ComposeStatus.ActiveIntercepts {
		// Filter by workspace name if specified
		if workspaceName != "" && intercept.WorkspaceName != workspaceName {
			continue
		}
		intercepts = append(intercepts, ActiveIntercept{
			EnvironmentName: env.Name,
			ServiceName:     intercept.ServiceName,
			Phase:           intercept.Phase,
			Message:         intercept.Message,
		})
	}

	return intercepts
}

func handleInterceptStop(serviceName string) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is connected to an environment
	if workspace.Status.ConnectedEnvironment == nil {
		return fmt.Errorf("workspace is not connected to any environment. Connect using 'kl env connect' first")
	}

	envName := workspace.Status.ConnectedEnvironment.Name
	if envName == "" {
		return fmt.Errorf("workspace has no connected environment")
	}

	// Get the connected environment
	env, err := getConnectedEnvironment(ctx, envName, workspace.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get environment '%s': %w", envName, err)
	}

	// Find the active intercept by service name
	_, err = findActiveInterceptByServiceName(ctx, serviceName, env, workspace.Name)
	if err != nil {
		return err
	}

	return handleInterceptStopWithEnv(ctx, env, serviceName, workspace.Namespace)
}

func handleInterceptStopWithEnv(ctx context.Context, env *environmentv1.Environment, serviceName, workspaceNamespace string) error {
	if env.Spec.Compose == nil {
		return fmt.Errorf("environment has no compose configuration")
	}

	// Find and disable/remove the intercept from compose spec
	found := false
	newIntercepts := []environmentv1.ServiceInterceptConfig{}
	for _, intercept := range env.Spec.Compose.Intercepts {
		if intercept.ServiceName == serviceName {
			found = true
			continue // Skip this intercept (remove it)
		}
		newIntercepts = append(newIntercepts, intercept)
	}

	if !found {
		return fmt.Errorf("no active intercept found for service '%s'", serviceName)
	}

	// Update environment spec
	env.Spec.Compose.Intercepts = newIntercepts

	if err := WsClient.K8sClient.Update(ctx, env); err != nil {
		return fmt.Errorf("failed to update environment: %w", err)
	}

	fmt.Printf("[✓] Service intercept removed\n")

	// Wait for the intercept to be deleted
	if err := waitForInterceptSync(ctx, env.Name, serviceName, workspaceNamespace, "stop"); err != nil {
		return fmt.Errorf("service intercept deletion failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("[✓] Service intercept has been removed\n")
	fmt.Printf("Service '%s' is no longer being intercepted\n", serviceName)
	fmt.Println("Normal traffic routing has been restored")

	return nil
}

func handleInterceptList() error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is connected to an environment
	if workspace.Status.ConnectedEnvironment == nil {
		fmt.Println("Workspace is not connected to any environment")
		return nil
	}

	envName := workspace.Status.ConnectedEnvironment.Name
	if envName == "" {
		fmt.Println("Workspace has no connected environment")
		return nil
	}

	// Get the connected environment
	env, err := getConnectedEnvironment(ctx, envName, workspace.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get environment '%s': %w", envName, err)
	}

	// Get all active intercepts (not filtered by workspace)
	intercepts := listActiveInterceptsFromEnv(env, "")

	if len(intercepts) == 0 {
		fmt.Println("No active service intercepts")
		fmt.Println("\nTo start intercepting a service, run:")
		fmt.Println("  kl intercept start")
		return nil
	}

	fmt.Printf("Active service intercepts (%d):\n\n", len(intercepts))
	for _, intercept := range intercepts {
		fmt.Printf("  %s\n", intercept.ServiceName)
		fmt.Printf("    Phase: %s\n", intercept.Phase)
		if intercept.Message != "" {
			fmt.Printf("    Message: %s\n", intercept.Message)
		}
		fmt.Println()
	}

	return nil
}

func handleInterceptStatus(serviceName string) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is connected to an environment
	if workspace.Status.ConnectedEnvironment == nil {
		return fmt.Errorf("workspace is not connected to any environment. Connect using 'kl env connect' first")
	}

	envName := workspace.Status.ConnectedEnvironment.Name
	if envName == "" {
		return fmt.Errorf("workspace has no connected environment")
	}

	// Get the connected environment
	env, err := getConnectedEnvironment(ctx, envName, workspace.Namespace)
	if err != nil {
		return fmt.Errorf("failed to get environment '%s': %w", envName, err)
	}

	if env.Spec.Compose == nil || env.Status.ComposeStatus == nil {
		fmt.Println("No active service intercepts")
		return nil
	}

	if serviceName != "" {
		// Show status for specific service
		_, err := findActiveInterceptByServiceName(ctx, serviceName, env, "")
		if err != nil {
			return err
		}

		// Find intercept in status
		var status *environmentv1.InterceptStatus
		for i, activeIntercept := range env.Status.ComposeStatus.ActiveIntercepts {
			if activeIntercept.ServiceName == serviceName {
				status = &env.Status.ComposeStatus.ActiveIntercepts[i]
				break
			}
		}

		// Find intercept in spec for port mappings
		var spec *environmentv1.ServiceInterceptConfig
		for i, specIntercept := range env.Spec.Compose.Intercepts {
			if specIntercept.ServiceName == serviceName {
				spec = &env.Spec.Compose.Intercepts[i]
				break
			}
		}

		if spec == nil && status == nil {
			return fmt.Errorf("no intercept found for service '%s'", serviceName)
		}

		printInterceptStatus(env.Name, serviceName, spec, status)
		return nil
	}

	// Show status for all intercepts
	first := true
	for _, status := range env.Status.ComposeStatus.ActiveIntercepts {
		if !first {
			fmt.Println("\n---")
		}
		first = false

		// Find matching spec
		var spec *environmentv1.ServiceInterceptConfig
		for i, intercept := range env.Spec.Compose.Intercepts {
			if intercept.ServiceName == status.ServiceName {
				spec = &env.Spec.Compose.Intercepts[i]
				break
			}
		}

		printInterceptStatus(env.Name, status.ServiceName, spec, &status)
	}

	if first {
		fmt.Println("No active service intercepts")
	}

	return nil
}

func printInterceptStatus(environmentName, serviceName string, spec *environmentv1.ServiceInterceptConfig, status *environmentv1.InterceptStatus) {
	fmt.Printf("Service: %s\n", serviceName)

	if status != nil {
		fmt.Printf("Phase: %s\n", status.Phase)
		if status.WorkspaceName != "" {
			fmt.Printf("Workspace: %s\n", status.WorkspaceName)
		}
		if status.Message != "" {
			fmt.Printf("Message: %s\n", status.Message)
		}
	} else {
		fmt.Printf("Phase: Pending\n")
	}

	if spec != nil && len(spec.PortMappings) > 0 {
		fmt.Println("\nPort Mappings:")
		for _, mapping := range spec.PortMappings {
			fmt.Printf("  %d (service) → %d (workspace)\n",
				mapping.ServicePort, mapping.WorkspacePort)
		}
	}
}

// waitForInterceptSync waits for the service intercept to sync to the desired state in environment's compose status
func waitForInterceptSync(ctx context.Context, envName, serviceName, workspaceNamespace, action string) error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Set up signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	if action == "start" {
		fmt.Print("Activating")
	} else {
		fmt.Print("Removing")
	}

	for {
		select {
		case <-sigChan:
			fmt.Println(" interrupted!")
			return fmt.Errorf("interrupted by user")
		case <-timeout:
			fmt.Println(" timeout!")
			return fmt.Errorf("timeout waiting for intercept sync after 30 seconds")
		case <-ticker.C:
			// Show progress
			fmt.Print(".")

			// Get current environment to check status
			env, err := getConnectedEnvironment(ctx, envName, workspaceNamespace)
			if err != nil {
				// Continue retrying on error
				continue
			}

			if env.Status.ComposeStatus == nil {
				continue // Status not yet populated
			}

			// Check if service is in activeIntercepts
			var found bool
			var interceptStatus environmentv1.InterceptStatus
			for _, activeIntercept := range env.Status.ComposeStatus.ActiveIntercepts {
				if activeIntercept.ServiceName == serviceName {
					found = true
					interceptStatus = activeIntercept
					break
				}
			}

			if action == "start" {
				// Wait for intercept to appear in activeIntercepts with active phase
				if !found {
					continue // Keep waiting for creation
				}

				// Check if intercept is active (only lowercase per CRD validation)
				if interceptStatus.Phase == "active" {
					fmt.Println(" done!")
					return nil
				}

				// Check if intercept failed (only lowercase per CRD validation)
				if interceptStatus.Phase == "failed" {
					fmt.Println(" failed!")
					message := interceptStatus.Message
					if message == "" {
						message = "unknown error"
					}
					return fmt.Errorf("service intercept failed: %s", message)
				}

				// Still creating, keep waiting
			} else {
				// action == "stop" - wait for intercept to be removed from activeIntercepts
				if !found {
					fmt.Println(" done!")
					return nil
				}

				// Intercept still exists, keep waiting
			}
		}
	}
}

func selectEnvironmentServiceWithFzf(services []EnvironmentService) (*EnvironmentService, error) {
	// Create input for fzf
	svcMap := make(map[string]*EnvironmentService)
	var items []string

	for i := range services {
		svc := &services[i]
		var line string
		if len(svc.Ports) > 0 {
			ports := make([]string, len(svc.Ports))
			for j, port := range svc.Ports {
				ports[j] = fmt.Sprintf("%d", port)
			}
			line = fmt.Sprintf("%s (%s) - ports: %s", svc.ServiceName, svc.State, strings.Join(ports, ", "))
		} else {
			line = fmt.Sprintf("%s (%s)", svc.ServiceName, svc.State)
		}
		items = append(items, line)
		svcMap[line] = svc
	}

	// Create temporary file for input
	tmpfile, err := os.CreateTemp("", "fzf-svc-*.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write items to temp file
	writer := bufio.NewWriter(tmpfile)
	for _, item := range items {
		fmt.Fprintln(writer, item)
	}
	writer.Flush()
	tmpfile.Close()

	// Open temp file for reading
	inputFile, err := os.Open(tmpfile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open temp file: %w", err)
	}
	defer inputFile.Close()

	// Save original stdin and replace it temporarily
	oldStdin := os.Stdin
	os.Stdin = inputFile
	defer func() {
		os.Stdin = oldStdin
		// Recover from any panics in fzf
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Error in fzf: %v\n", r)
		}
	}()

	// Parse fzf options with embedded fzf
	opts, err := fzf.ParseOptions(true, []string{
		"--height=40%",
		"--layout=reverse",
		"--border",
		"--prompt=Select service to intercept: ",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse fzf options: %w", err)
	}

	// Printer function to capture output
	var selected string
	opts.Printer = func(s string) {
		selected = s
	}

	// Run fzf
	exitCode, err := fzf.Run(opts)

	if exitCode != fzf.ExitOk || err != nil {
		return nil, fmt.Errorf("service selection cancelled")
	}

	if selected == "" {
		return nil, fmt.Errorf("no service selected")
	}

	svc, ok := svcMap[selected]
	if !ok {
		return nil, fmt.Errorf("invalid service selected")
	}

	return svc, nil
}

func selectActiveInterceptWithFzf(intercepts []ActiveIntercept) (*ActiveIntercept, error) {
	// Create input for fzf
	interceptMap := make(map[string]*ActiveIntercept)
	var items []string

	for i := range intercepts {
		intercept := &intercepts[i]
		line := fmt.Sprintf("%s (%s)", intercept.ServiceName, intercept.Phase)
		items = append(items, line)
		interceptMap[line] = intercept
	}

	// Create temporary file for input
	tmpfile, err := os.CreateTemp("", "fzf-intercept-*.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write items to temp file
	writer := bufio.NewWriter(tmpfile)
	for _, item := range items {
		fmt.Fprintln(writer, item)
	}
	writer.Flush()
	tmpfile.Close()

	// Open temp file for reading
	inputFile, err := os.Open(tmpfile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open temp file: %w", err)
	}
	defer inputFile.Close()

	// Save original stdin and replace it temporarily
	oldStdin := os.Stdin
	os.Stdin = inputFile
	defer func() {
		os.Stdin = oldStdin
		// Recover from any panics in fzf
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Error in fzf: %v\n", r)
		}
	}()

	// Parse fzf options with embedded fzf
	opts, err := fzf.ParseOptions(true, []string{
		"--height=40%",
		"--layout=reverse",
		"--border",
		"--prompt=Select intercept to stop: ",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse fzf options: %w", err)
	}

	// Printer function to capture output
	var selected string
	opts.Printer = func(s string) {
		selected = s
	}

	// Run fzf
	exitCode, err := fzf.Run(opts)

	if exitCode != fzf.ExitOk || err != nil {
		return nil, fmt.Errorf("intercept selection cancelled")
	}

	if selected == "" {
		return nil, fmt.Errorf("no intercept selected")
	}

	intercept, ok := interceptMap[selected]
	if !ok {
		return nil, fmt.Errorf("invalid intercept selected")
	}

	return intercept, nil
}

// getAvailableServiceNames returns a list of available service names for shell completion
func getAvailableServiceNames() []string {
	if err := InitClient(); err != nil {
		return nil
	}
	ctx := context.Background()

	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return nil
	}

	if workspace.Status.ConnectedEnvironment == nil {
		return nil
	}

	envName := workspace.Status.ConnectedEnvironment.Name
	if envName == "" {
		return nil
	}

	env, err := getConnectedEnvironment(ctx, envName, workspace.Namespace)
	if err != nil {
		return nil
	}

	services, err := listEnvironmentServices(ctx, env)
	if err != nil {
		return nil
	}

	var names []string
	for _, svc := range services {
		names = append(names, svc.ServiceName)
	}
	return names
}

// getActiveInterceptNames returns a list of active intercept service names for shell completion
func getActiveInterceptNames() []string {
	if err := InitClient(); err != nil {
		return nil
	}
	ctx := context.Background()

	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return nil
	}

	if workspace.Status.ConnectedEnvironment == nil {
		return nil
	}

	envName := workspace.Status.ConnectedEnvironment.Name
	if envName == "" {
		return nil
	}

	env, err := getConnectedEnvironment(ctx, envName, workspace.Namespace)
	if err != nil {
		return nil
	}

	intercepts := listActiveInterceptsFromEnv(env, workspace.Name)

	var names []string
	for _, intercept := range intercepts {
		names = append(names, intercept.ServiceName)
	}
	return names
}

// selectPortsWithFzf allows user to select which ports to intercept using fzf multi-select
func selectPortsWithFzf(serviceName string, ports []int32) ([]int32, error) {
	// If only one port, auto-select it
	if len(ports) == 1 {
		fmt.Printf("\nService '%s' has 1 port: %d (auto-selected)\n", serviceName, ports[0])
		return ports, nil
	}

	// Create input for fzf
	portMap := make(map[string]int32)
	var items []string

	for _, port := range ports {
		line := fmt.Sprintf("%d", port)
		items = append(items, line)
		portMap[line] = port
	}

	// Create temporary file for input
	tmpfile, err := os.CreateTemp("", "fzf-ports-*.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write items to temp file
	writer := bufio.NewWriter(tmpfile)
	for _, item := range items {
		fmt.Fprintln(writer, item)
	}
	writer.Flush()
	tmpfile.Close()

	// Open temp file for reading
	inputFile, err := os.Open(tmpfile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open temp file: %w", err)
	}
	defer inputFile.Close()

	// Save original stdin and replace it temporarily
	oldStdin := os.Stdin
	os.Stdin = inputFile
	defer func() {
		os.Stdin = oldStdin
		// Recover from any panics in fzf
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Error in fzf: %v\n", r)
		}
	}()

	fmt.Printf("\nService '%s' has %d ports. Select ports to intercept (TAB to select, ENTER to confirm):\n", serviceName, len(ports))

	// Parse fzf options with multi-select enabled
	opts, err := fzf.ParseOptions(true, []string{
		"--height=40%",
		"--layout=reverse",
		"--border",
		"--multi",
		"--prompt=Select ports (TAB to toggle): ",
		"--header=Press TAB to select/deselect, ENTER to confirm",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse fzf options: %w", err)
	}

	// Collect all selected items
	var selectedItems []string
	opts.Printer = func(s string) {
		selectedItems = append(selectedItems, s)
	}

	// Run fzf
	exitCode, err := fzf.Run(opts)

	if exitCode != fzf.ExitOk || err != nil {
		return nil, fmt.Errorf("port selection cancelled")
	}

	if len(selectedItems) == 0 {
		return nil, fmt.Errorf("no ports selected")
	}

	// Convert selected items back to port numbers
	var selectedPorts []int32
	for _, item := range selectedItems {
		if port, ok := portMap[item]; ok {
			selectedPorts = append(selectedPorts, port)
		}
	}

	return selectedPorts, nil
}
