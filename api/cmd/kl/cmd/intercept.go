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

  # Start intercepting a specific service (composition/service format)
  kl intercept start myapp/api-server
  kl i s myapp/api-server

  # List active intercepts
  kl intercept list
  kl i ls

  # Stop intercepting a service
  kl intercept stop myapp/api-server
  kl i sp myapp/api-server

  # Show intercept status
  kl intercept status
  kl i st`,
}

var interceptStartCmd = &cobra.Command{
	Use:     "start [composition/service-name]",
	Aliases: []string{"s", "begin"},
	Short:   "Start intercepting a service",
	Long: `Start intercepting a service to redirect its traffic to your workspace.

If no service name is provided, an interactive list will be shown.
You will be prompted to map each service port to a workspace port.

Services are specified in the format: composition/service-name`,
	Example: `  # Interactive service selection
  kl intercept start
  kl i s

  # Intercept specific service
  kl intercept start myapp/api-server
  kl i s myapp/api-server`,
	Args: cobra.MaximumNArgs(1),
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
	Use:     "stop [composition/service-name]",
	Aliases: []string{"sp", "end"},
	Short:   "Stop intercepting a service",
	Long: `Stop intercepting a service and restore normal traffic routing.

If no service name is provided, an interactive list of active intercepts will be shown.

Services are specified in the format: composition/service-name`,
	Example: `  # Interactive intercept selection
  kl intercept stop
  kl i sp

  # Stop specific service intercept
  kl intercept stop myapp/api-server
  kl i sp myapp/api-server`,
	Args: cobra.MaximumNArgs(1),
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
	Use:     "status [composition/service-name]",
	Aliases: []string{"st"},
	Short:   "Show intercept status",
	Long:    `Show detailed status of service intercept(s).`,
	Example: `  # Show all intercepts status
  kl intercept status
  kl i st

  # Show specific service intercept status
  kl intercept status myapp/api-server
  kl i st myapp/api-server`,
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

// CompositionService represents a service from a composition
type CompositionService struct {
	CompositionName string
	ServiceName     string
	Ports           []int32
	State           string
}

// listCompositionServices lists all services from all compositions in the target namespace
func listCompositionServices(ctx context.Context, targetNamespace string) ([]CompositionService, error) {
	compList := &environmentv1.CompositionList{}
	if err := WsClient.K8sClient.List(ctx, compList, client.InNamespace(targetNamespace)); err != nil {
		return nil, fmt.Errorf("failed to list compositions: %w", err)
	}

	var services []CompositionService
	for _, comp := range compList.Items {
		for _, svcStatus := range comp.Status.Services {
			services = append(services, CompositionService{
				CompositionName: comp.Name,
				ServiceName:     svcStatus.Name,
				Ports:           svcStatus.Ports,
				State:           svcStatus.State,
			})
		}
	}

	return services, nil
}

// getComposition gets a composition by name in the target namespace
func getComposition(ctx context.Context, compositionName, targetNamespace string) (*environmentv1.Composition, error) {
	comp := &environmentv1.Composition{}
	if err := WsClient.K8sClient.Get(ctx, client.ObjectKey{
		Name:      compositionName,
		Namespace: targetNamespace,
	}, comp); err != nil {
		return nil, err
	}
	return comp, nil
}

// parseServiceRef parses "composition/service" format into composition and service names
func parseServiceRef(serviceRef string) (compositionName, serviceName string, err error) {
	parts := strings.SplitN(serviceRef, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid service format. Use 'composition/service-name' format")
	}
	return parts[0], parts[1], nil
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

	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace
	if targetNamespace == "" {
		return fmt.Errorf("workspace has no connected environment namespace")
	}

	// List services from compositions
	fmt.Println("Loading services...")
	services, err := listCompositionServices(ctx, targetNamespace)
	if err != nil {
		return err
	}

	if len(services) == 0 {
		return fmt.Errorf("no services found in compositions")
	}

	// Use embedded fzf to select service
	selectedService, err := selectCompositionServiceWithFzf(services)
	if err != nil {
		return err
	}

	return handleInterceptStartWithCompositionService(ctx, *selectedService, workspace.Name, workspace.Namespace, targetNamespace)
}

func handleInterceptStart(serviceRef string) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Parse composition/service format
	compositionName, serviceName, err := parseServiceRef(serviceRef)
	if err != nil {
		return err
	}

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is connected to an environment
	if workspace.Status.ConnectedEnvironment == nil {
		return fmt.Errorf("workspace is not connected to any environment. Connect using 'kl env connect' first")
	}

	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace
	if targetNamespace == "" {
		return fmt.Errorf("workspace has no connected environment namespace")
	}

	// Get the composition
	comp, err := getComposition(ctx, compositionName, targetNamespace)
	if err != nil {
		return fmt.Errorf("failed to get composition '%s': %w", compositionName, err)
	}

	// Find the service in composition status
	var svc *CompositionService
	for _, svcStatus := range comp.Status.Services {
		if svcStatus.Name == serviceName {
			svc = &CompositionService{
				CompositionName: compositionName,
				ServiceName:     svcStatus.Name,
				Ports:           svcStatus.Ports,
				State:           svcStatus.State,
			}
			break
		}
	}

	if svc == nil {
		return fmt.Errorf("service '%s' not found in composition '%s'", serviceName, compositionName)
	}

	return handleInterceptStartWithCompositionService(ctx, *svc, workspace.Name, workspace.Namespace, targetNamespace)
}

func handleInterceptStartWithCompositionService(ctx context.Context, svc CompositionService, workspaceName, workspaceNamespace, targetNamespace string) error {
	// Get the composition
	comp, err := getComposition(ctx, svc.CompositionName, targetNamespace)
	if err != nil {
		return fmt.Errorf("failed to get composition '%s': %w", svc.CompositionName, err)
	}

	// Check if intercept already exists in composition spec
	for _, intercept := range comp.Spec.Intercepts {
		if intercept.ServiceName == svc.ServiceName && intercept.Enabled {
			return fmt.Errorf("service '%s/%s' is already being intercepted", svc.CompositionName, svc.ServiceName)
		}
	}

	// Prompt for port mappings
	var portMappings []environmentv1.PortMapping
	reader := bufio.NewReader(os.Stdin)

	if len(svc.Ports) > 0 {
		// Service has defined ports
		fmt.Printf("\nService '%s/%s' has %d port(s):\n", svc.CompositionName, svc.ServiceName, len(svc.Ports))
		portMappings = make([]environmentv1.PortMapping, 0, len(svc.Ports))

		for _, port := range svc.Ports {
			fmt.Printf("\n  Service port: %d\n", port)
			fmt.Printf("  Workspace port [%d]: ", port)
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
		fmt.Printf("\nService '%s/%s' has no ports defined.\n", svc.CompositionName, svc.ServiceName)
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
	for i, existing := range comp.Spec.Intercepts {
		if existing.ServiceName == svc.ServiceName {
			comp.Spec.Intercepts[i] = interceptConfig
			found = true
			break
		}
	}
	if !found {
		comp.Spec.Intercepts = append(comp.Spec.Intercepts, interceptConfig)
	}

	// Update composition
	if err := WsClient.K8sClient.Update(ctx, comp); err != nil {
		return fmt.Errorf("failed to update composition: %w", err)
	}

	fmt.Printf("\n[✓] Service intercept added\n")

	// Wait for the intercept to become active
	if err := waitForInterceptSync(svc.CompositionName, svc.ServiceName, targetNamespace, "start"); err != nil {
		return fmt.Errorf("service intercept activation failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("[✓] Service intercept is now active\n")
	fmt.Printf("Service '%s/%s' is being intercepted\n\n", svc.CompositionName, svc.ServiceName)
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

	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace
	if targetNamespace == "" {
		return fmt.Errorf("workspace has no connected environment namespace")
	}

	// Get all active intercepts from compositions
	activeIntercepts, err := listActiveIntercepts(ctx, targetNamespace, workspace.Name)
	if err != nil {
		return err
	}

	if len(activeIntercepts) == 0 {
		return fmt.Errorf("no active service intercepts found")
	}

	// Use embedded fzf to select intercept
	selectedIntercept, err := selectActiveInterceptWithFzf(activeIntercepts)
	if err != nil {
		return err
	}

	return handleInterceptStopWithRef(ctx, selectedIntercept.CompositionName, selectedIntercept.ServiceName, targetNamespace)
}

// ActiveIntercept represents an active intercept for display
type ActiveIntercept struct {
	CompositionName string
	ServiceName     string
	Phase           string
	Message         string
}

// listActiveIntercepts lists all active intercepts for the given workspace
func listActiveIntercepts(ctx context.Context, targetNamespace, workspaceName string) ([]ActiveIntercept, error) {
	compList := &environmentv1.CompositionList{}
	if err := WsClient.K8sClient.List(ctx, compList, client.InNamespace(targetNamespace)); err != nil {
		return nil, fmt.Errorf("failed to list compositions: %w", err)
	}

	var intercepts []ActiveIntercept
	for _, comp := range compList.Items {
		for _, intercept := range comp.Status.ActiveIntercepts {
			// Filter by workspace name if specified
			if workspaceName != "" && intercept.WorkspaceName != workspaceName {
				continue
			}
			intercepts = append(intercepts, ActiveIntercept{
				CompositionName: comp.Name,
				ServiceName:     intercept.ServiceName,
				Phase:           intercept.Phase,
				Message:         intercept.Message,
			})
		}
	}

	return intercepts, nil
}

func handleInterceptStop(serviceRef string) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Parse composition/service format
	compositionName, serviceName, err := parseServiceRef(serviceRef)
	if err != nil {
		return err
	}

	// Get current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace is connected to an environment
	if workspace.Status.ConnectedEnvironment == nil {
		return fmt.Errorf("workspace is not connected to any environment. Connect using 'kl env connect' first")
	}

	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace

	return handleInterceptStopWithRef(ctx, compositionName, serviceName, targetNamespace)
}

func handleInterceptStopWithRef(ctx context.Context, compositionName, serviceName, targetNamespace string) error {
	// Get the composition
	comp, err := getComposition(ctx, compositionName, targetNamespace)
	if err != nil {
		return fmt.Errorf("failed to get composition '%s': %w", compositionName, err)
	}

	// Find and disable/remove the intercept from composition spec
	found := false
	newIntercepts := []environmentv1.ServiceInterceptConfig{}
	for _, intercept := range comp.Spec.Intercepts {
		if intercept.ServiceName == serviceName {
			found = true
			continue // Skip this intercept (remove it)
		}
		newIntercepts = append(newIntercepts, intercept)
	}

	if !found {
		return fmt.Errorf("no active intercept found for service '%s/%s'", compositionName, serviceName)
	}

	// Update composition spec
	comp.Spec.Intercepts = newIntercepts

	if err := WsClient.K8sClient.Update(ctx, comp); err != nil {
		return fmt.Errorf("failed to update composition: %w", err)
	}

	fmt.Printf("[✓] Service intercept removed\n")

	// Wait for the intercept to be deleted
	if err := waitForInterceptSync(compositionName, serviceName, targetNamespace, "stop"); err != nil {
		return fmt.Errorf("service intercept deletion failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("[✓] Service intercept has been removed\n")
	fmt.Printf("Service '%s/%s' is no longer being intercepted\n", compositionName, serviceName)
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

	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace

	// Get all active intercepts (not filtered by workspace)
	intercepts, err := listActiveIntercepts(ctx, targetNamespace, "")
	if err != nil {
		return err
	}

	if len(intercepts) == 0 {
		fmt.Println("No active service intercepts")
		fmt.Println("\nTo start intercepting a service, run:")
		fmt.Println("  kl intercept start")
		return nil
	}

	fmt.Printf("Active service intercepts (%d):\n\n", len(intercepts))
	for _, intercept := range intercepts {
		fmt.Printf("  %s/%s\n", intercept.CompositionName, intercept.ServiceName)
		fmt.Printf("    Phase: %s\n", intercept.Phase)
		if intercept.Message != "" {
			fmt.Printf("    Message: %s\n", intercept.Message)
		}
		fmt.Println()
	}

	return nil
}

func handleInterceptStatus(serviceRef string) error {
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

	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace

	if serviceRef != "" {
		// Show status for specific service
		compositionName, serviceName, err := parseServiceRef(serviceRef)
		if err != nil {
			return err
		}

		comp, err := getComposition(ctx, compositionName, targetNamespace)
		if err != nil {
			return fmt.Errorf("failed to get composition '%s': %w", compositionName, err)
		}

		// Find intercept in status
		var status *environmentv1.InterceptStatus
		for i, intercept := range comp.Status.ActiveIntercepts {
			if intercept.ServiceName == serviceName {
				status = &comp.Status.ActiveIntercepts[i]
				break
			}
		}

		// Find intercept in spec for port mappings
		var spec *environmentv1.ServiceInterceptConfig
		for i, intercept := range comp.Spec.Intercepts {
			if intercept.ServiceName == serviceName {
				spec = &comp.Spec.Intercepts[i]
				break
			}
		}

		if spec == nil && status == nil {
			return fmt.Errorf("no intercept found for service '%s/%s'", compositionName, serviceName)
		}

		printInterceptStatus(compositionName, serviceName, spec, status)
		return nil
	}

	// Show status for all intercepts
	compList := &environmentv1.CompositionList{}
	if err := WsClient.K8sClient.List(ctx, compList, client.InNamespace(targetNamespace)); err != nil {
		return fmt.Errorf("failed to list compositions: %w", err)
	}

	first := true
	for _, comp := range compList.Items {
		for _, status := range comp.Status.ActiveIntercepts {
			if !first {
				fmt.Println("\n---")
			}
			first = false

			// Find matching spec
			var spec *environmentv1.ServiceInterceptConfig
			for i, intercept := range comp.Spec.Intercepts {
				if intercept.ServiceName == status.ServiceName {
					spec = &comp.Spec.Intercepts[i]
					break
				}
			}

			printInterceptStatus(comp.Name, status.ServiceName, spec, &status)
		}
	}

	if first {
		fmt.Println("No active service intercepts")
	}

	return nil
}

func printInterceptStatus(compositionName, serviceName string, spec *environmentv1.ServiceInterceptConfig, status *environmentv1.InterceptStatus) {
	fmt.Printf("Service: %s/%s\n", compositionName, serviceName)

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

// waitForInterceptSync waits for the service intercept to sync to the desired state in composition status
func waitForInterceptSync(compositionName, serviceName, targetNamespace, action string) error {
	ctx := context.Background()
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

			// Get current composition to check status
			comp, err := getComposition(ctx, compositionName, targetNamespace)
			if err != nil {
				// Continue retrying on error
				continue
			}

			// Check if service is in activeIntercepts
			var found bool
			var interceptStatus environmentv1.InterceptStatus
			for _, activeIntercept := range comp.Status.ActiveIntercepts {
				if activeIntercept.ServiceName == serviceName {
					found = true
					interceptStatus = activeIntercept
					break
				}
			}

			if action == "start" {
				// Wait for intercept to appear in activeIntercepts with Active phase
				if !found {
					continue // Keep waiting for creation
				}

				// Check if intercept is active
				if interceptStatus.Phase == "active" || interceptStatus.Phase == "Active" {
					fmt.Println(" done!")
					return nil
				}

				// Check if intercept failed
				if interceptStatus.Phase == "failed" || interceptStatus.Phase == "Failed" {
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

func selectCompositionServiceWithFzf(services []CompositionService) (*CompositionService, error) {
	// Create input for fzf
	svcMap := make(map[string]*CompositionService)
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
		line := fmt.Sprintf("%s/%s (%s)", intercept.CompositionName, intercept.ServiceName, intercept.Phase)
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
