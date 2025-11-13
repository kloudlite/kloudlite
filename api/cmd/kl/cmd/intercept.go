package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	fzf "github.com/junegunn/fzf/src"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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

	// List services in the target namespace
	fmt.Printf("Loading services from namespace '%s'...\n", targetNamespace)
	serviceList := &corev1.ServiceList{}
	if err := WsClient.K8sClient.List(ctx, serviceList, client.InNamespace(targetNamespace)); err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	if len(serviceList.Items) == 0 {
		return fmt.Errorf("no services found in namespace '%s'", targetNamespace)
	}

	// Use embedded fzf to select service
	selectedService, err := selectServiceWithFzf(serviceList.Items)
	if err != nil {
		return err
	}

	return handleInterceptStartWithService(*selectedService, workspace, targetNamespace)
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

	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace
	if targetNamespace == "" {
		return fmt.Errorf("workspace has no connected environment namespace")
	}

	// Get the service
	service := &corev1.Service{}
	if err := WsClient.K8sClient.Get(ctx, types.NamespacedName{
		Name:      serviceName,
		Namespace: targetNamespace,
	}, service); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("service '%s' not found in namespace '%s'", serviceName, targetNamespace)
		}
		return fmt.Errorf("failed to get service: %w", err)
	}

	return handleInterceptStartWithService(*service, workspace, targetNamespace)
}

func handleInterceptStartWithService(service corev1.Service, workspace *workspacesv1.Workspace, targetNamespace string) error {
	ctx := context.Background()

	// Check if environment connection exists
	if workspace.Spec.EnvironmentConnection == nil {
		return fmt.Errorf("workspace is not connected to any environment. Connect using 'kl env connect' first")
	}

	// Check if intercept already exists in workspace spec
	for _, intercept := range workspace.Spec.EnvironmentConnection.Intercepts {
		if intercept.ServiceName == service.Name {
			return fmt.Errorf("service '%s' is already being intercepted by this workspace", service.Name)
		}
	}

	// Prompt for port mappings
	var portMappings []workspacesv1.PortMapping
	reader := bufio.NewReader(os.Stdin)

	if len(service.Spec.Ports) > 0 {
		// Service has defined ports
		fmt.Printf("\nService '%s' has %d port(s):\n", service.Name, len(service.Spec.Ports))
		portMappings = make([]workspacesv1.PortMapping, 0, len(service.Spec.Ports))

		for _, port := range service.Spec.Ports {
			fmt.Printf("\n  Service port: %d/%s\n", port.Port, port.Protocol)
			fmt.Printf("  Workspace port [%d]: ", port.Port)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			input = strings.TrimSpace(input)
			workspacePort := port.Port
			if input != "" {
				parsedPort, err := strconv.Atoi(input)
				if err != nil {
					return fmt.Errorf("invalid port number: %s", input)
				}
				workspacePort = int32(parsedPort)
			}

			portMappings = append(portMappings, workspacesv1.PortMapping{
				ServicePort:   port.Port,
				WorkspacePort: workspacePort,
				Protocol:      port.Protocol,
			})
		}
	} else {
		// Portless service (e.g., headless service) - ask user to manually configure ports
		fmt.Printf("\nService '%s' is a portless service (headless service).\n", service.Name)
		fmt.Println("You need to manually specify port mappings.")
		fmt.Println("Enter port mappings (press Enter with empty service port to finish):")

		portMappings = make([]workspacesv1.PortMapping, 0)

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

			fmt.Print("  Protocol (TCP/UDP) [TCP]: ")
			protocolInput, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}

			protocol := strings.ToUpper(strings.TrimSpace(protocolInput))
			if protocol == "" {
				protocol = "TCP"
			}

			if protocol != "TCP" && protocol != "UDP" && protocol != "SCTP" {
				fmt.Println("  Invalid protocol. Using TCP as default.")
				protocol = "TCP"
			}

			portMappings = append(portMappings, workspacesv1.PortMapping{
				ServicePort:   int32(servicePort),
				WorkspacePort: int32(workspacePort),
				Protocol:      corev1.Protocol(protocol),
			})

			fmt.Printf("  ✓ Added mapping: %d (service) → %d (workspace) [%s]\n\n", servicePort, workspacePort, protocol)
		}
	}

	// Add intercept to workspace spec
	interceptSpec := workspacesv1.InterceptSpec{
		ServiceName:  service.Name,
		PortMappings: portMappings,
	}

	if workspace.Spec.EnvironmentConnection.Intercepts == nil {
		workspace.Spec.EnvironmentConnection.Intercepts = []workspacesv1.InterceptSpec{}
	}
	workspace.Spec.EnvironmentConnection.Intercepts = append(workspace.Spec.EnvironmentConnection.Intercepts, interceptSpec)

	// Update workspace
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("\n[✓] Service intercept added to workspace spec\n")

	// Wait for the intercept to become active
	if err := waitForInterceptSync(service.Name, targetNamespace, "start"); err != nil {
		return fmt.Errorf("service intercept activation failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("[✓] Service intercept is now active\n")
	fmt.Printf("Service '%s' is being intercepted by workspace '%s'\n\n", service.Name, workspace.Name)
	fmt.Println("Port mappings:")
	for _, mapping := range portMappings {
		fmt.Printf("  %d (service) → %d (workspace)\n", mapping.ServicePort, mapping.WorkspacePort)
	}
	fmt.Printf("\nTraffic to %s.%s.svc.cluster.local is now routed to your workspace.\n", service.Name, targetNamespace)

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
	if workspace.Spec.EnvironmentConnection == nil {
		return fmt.Errorf("workspace is not connected to any environment. Connect using 'kl env connect' first")
	}

	if len(workspace.Spec.EnvironmentConnection.Intercepts) == 0 {
		return fmt.Errorf("no active service intercepts found")
	}

	// Use embedded fzf to select intercept from workspace spec
	selectedServiceName, err := selectInterceptFromWorkspaceSpec(workspace.Spec.EnvironmentConnection.Intercepts)
	if err != nil {
		return err
	}

	return handleInterceptStop(selectedServiceName)
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
	if workspace.Spec.EnvironmentConnection == nil {
		return fmt.Errorf("workspace is not connected to any environment. Connect using 'kl env connect' first")
	}

	// Find and remove the intercept from workspace spec
	found := false
	newIntercepts := []workspacesv1.InterceptSpec{}
	for _, intercept := range workspace.Spec.EnvironmentConnection.Intercepts {
		if intercept.ServiceName == serviceName {
			found = true
			continue // Skip this intercept (remove it)
		}
		newIntercepts = append(newIntercepts, intercept)
	}

	if !found {
		return fmt.Errorf("no active intercept found for service '%s'", serviceName)
	}

	// Get target namespace before updating
	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace

	// Update workspace spec
	workspace.Spec.EnvironmentConnection.Intercepts = newIntercepts

	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("[✓] Service intercept removed from workspace spec\n")

	// Wait for the intercept to be deleted
	if err := waitForInterceptSync(serviceName, targetNamespace, "stop"); err != nil {
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
	if workspace.Spec.EnvironmentConnection == nil {
		fmt.Println("Workspace is not connected to any environment")
		return nil
	}

	if len(workspace.Spec.EnvironmentConnection.Intercepts) == 0 {
		fmt.Println("No service intercepts configured")
		fmt.Println("\nTo start intercepting a service, run:")
		fmt.Println("  kl intercept start")
		return nil
	}

	// Get target namespace from status
	targetNamespace := ""
	if workspace.Status.ConnectedEnvironment != nil {
		targetNamespace = workspace.Status.ConnectedEnvironment.TargetNamespace
	}

	// Get status from workspace status (intercepts are now tracked in workspace.status.activeIntercepts)
	activeInterceptsMap := make(map[string]workspacesv1.InterceptStatus)
	for _, activeIntercept := range workspace.Status.ActiveIntercepts {
		activeInterceptsMap[activeIntercept.ServiceName] = activeIntercept
	}

	fmt.Printf("Service intercepts (%d):\n\n", len(workspace.Spec.EnvironmentConnection.Intercepts))
	for _, interceptSpec := range workspace.Spec.EnvironmentConnection.Intercepts {
		phase := "Pending"
		message := ""

		// Get status from workspace.status.activeIntercepts
		if status, exists := activeInterceptsMap[interceptSpec.ServiceName]; exists {
			phase = status.Phase
			message = status.Message
		}

		fmt.Printf("  Service: %s\n", interceptSpec.ServiceName)
		fmt.Printf("  Phase: %s\n", phase)
		fmt.Printf("  Port mappings:\n")
		for _, mapping := range interceptSpec.PortMappings {
			fmt.Printf("    %d → %d (%s)\n", mapping.ServicePort, mapping.WorkspacePort, mapping.Protocol)
		}
		if message != "" {
			fmt.Printf("  Message: %s\n", message)
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

	if workspace.Spec.EnvironmentConnection == nil || len(workspace.Spec.EnvironmentConnection.Intercepts) == 0 {
		fmt.Println("No service intercepts configured")
		return nil
	}

	// Build map of active intercept statuses
	activeInterceptsMap := make(map[string]workspacesv1.InterceptStatus)
	for _, activeIntercept := range workspace.Status.ActiveIntercepts {
		activeInterceptsMap[activeIntercept.ServiceName] = activeIntercept
	}

	if serviceName != "" {
		// Show status for specific service
		var found bool
		var interceptSpec workspacesv1.InterceptSpec
		for _, spec := range workspace.Spec.EnvironmentConnection.Intercepts {
			if spec.ServiceName == serviceName {
				found = true
				interceptSpec = spec
				break
			}
		}

		if !found {
			return fmt.Errorf("no intercept configured for service '%s'", serviceName)
		}

		status, hasStatus := activeInterceptsMap[serviceName]
		printInterceptStatusFromWorkspace(interceptSpec, status, hasStatus, workspace.Name)
		return nil
	}

	// Show status for all intercepts
	for i, interceptSpec := range workspace.Spec.EnvironmentConnection.Intercepts {
		if i > 0 {
			fmt.Println("\n---\n")
		}
		status, hasStatus := activeInterceptsMap[interceptSpec.ServiceName]
		printInterceptStatusFromWorkspace(interceptSpec, status, hasStatus, workspace.Name)
	}

	return nil
}

func printInterceptStatusFromWorkspace(interceptSpec workspacesv1.InterceptSpec, status workspacesv1.InterceptStatus, hasStatus bool, workspaceName string) {
	fmt.Printf("Service: %s\n", interceptSpec.ServiceName)
	fmt.Printf("Workspace: %s\n", workspaceName)

	phase := "Pending"
	if hasStatus {
		phase = status.Phase
	}
	fmt.Printf("Phase: %s\n", phase)

	if hasStatus && status.Message != "" {
		fmt.Printf("Message: %s\n", status.Message)
	}

	fmt.Println("\nPort Mappings:")
	for _, mapping := range interceptSpec.PortMappings {
		fmt.Printf("  %d (service) → %d (workspace) [%s]\n",
			mapping.ServicePort, mapping.WorkspacePort, mapping.Protocol)
	}

	if hasStatus && status.InterceptPodName != "" {
		fmt.Printf("\nIntercept Pod: %s\n", status.InterceptPodName)
	}
}

// waitForInterceptSync waits for the service intercept to sync to the desired state in workspace status
func waitForInterceptSync(serviceName, targetNamespace, action string) error {
	ctx := context.Background()
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	if action == "start" {
		fmt.Print("Waiting for service intercept to become active")
	} else {
		fmt.Print("Waiting for service intercept to be removed")
	}

	for {
		select {
		case <-timeout:
			fmt.Println(" timeout!")
			return fmt.Errorf("timeout waiting for intercept sync after 30 seconds")
		case <-ticker.C:
			// Show progress
			fmt.Print(".")

			// Get current workspace to check status
			workspace, err := WsClient.Get(ctx)
			if err != nil {
				// Continue retrying on error
				continue
			}

			// Check if service is in activeIntercepts
			var found bool
			var interceptStatus workspacesv1.InterceptStatus
			for _, activeIntercept := range workspace.Status.ActiveIntercepts {
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
				if interceptStatus.Phase == "Active" {
					fmt.Println(" done!")
					return nil
				}

				// Check if intercept failed
				if interceptStatus.Phase == "Failed" {
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

func selectServiceWithFzf(services []corev1.Service) (*corev1.Service, error) {
	// Create input for fzf
	svcMap := make(map[string]*corev1.Service)
	var items []string

	for i := range services {
		svc := &services[i]
		portStr := ""
		if len(svc.Spec.Ports) > 0 {
			ports := make([]string, len(svc.Spec.Ports))
			for j, port := range svc.Spec.Ports {
				ports[j] = fmt.Sprintf("%d/%s", port.Port, port.Protocol)
			}
			portStr = strings.Join(ports, ", ")
		}
		line := fmt.Sprintf("%s (%s) - %s", svc.Name, svc.Spec.Type, portStr)
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
		"--prompt=Select service: ",
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

func selectInterceptFromWorkspaceSpec(intercepts []workspacesv1.InterceptSpec) (string, error) {
	// Create input for fzf
	serviceMap := make(map[string]string)
	var items []string

	for _, interceptSpec := range intercepts {
		portStr := ""
		if len(interceptSpec.PortMappings) > 0 {
			ports := make([]string, len(interceptSpec.PortMappings))
			for j, mapping := range interceptSpec.PortMappings {
				ports[j] = fmt.Sprintf("%d→%d", mapping.ServicePort, mapping.WorkspacePort)
			}
			portStr = strings.Join(ports, ", ")
		}

		line := fmt.Sprintf("%s - %s", interceptSpec.ServiceName, portStr)
		items = append(items, line)
		serviceMap[line] = interceptSpec.ServiceName
	}

	// Create temporary file for input
	tmpfile, err := os.CreateTemp("", "fzf-intercept-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
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
		return "", fmt.Errorf("failed to open temp file: %w", err)
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
		return "", fmt.Errorf("failed to parse fzf options: %w", err)
	}

	// Printer function to capture output
	var selected string
	opts.Printer = func(s string) {
		selected = s
	}

	// Run fzf
	exitCode, err := fzf.Run(opts)

	if exitCode != fzf.ExitOk || err != nil {
		return "", fmt.Errorf("intercept selection cancelled")
	}

	if selected == "" {
		return "", fmt.Errorf("no intercept selected")
	}

	serviceName, ok := serviceMap[selected]
	if !ok {
		return "", fmt.Errorf("invalid intercept selected")
	}

	return serviceName, nil
}
