package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	fzf "github.com/junegunn/fzf/src"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// Check if intercept already exists
	interceptName := fmt.Sprintf("%s-%s", service.Name, workspace.Name)
	existingIntercept := &interceptsv1.ServiceIntercept{}
	err := WsClient.K8sClient.Get(ctx, types.NamespacedName{
		Name:      interceptName,
		Namespace: targetNamespace,
	}, existingIntercept)
	if err == nil {
		return fmt.Errorf("service '%s' is already being intercepted by this workspace", service.Name)
	} else if !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to check existing intercept: %w", err)
	}

	// Prompt for port mappings
	fmt.Printf("\nService '%s' has %d port(s):\n", service.Name, len(service.Spec.Ports))
	portMappings := make([]interceptsv1.PortMapping, 0, len(service.Spec.Ports))

	reader := bufio.NewReader(os.Stdin)
	for _, port := range service.Spec.Ports {
		fmt.Printf("  Port %d/%s -> Workspace port [%d]: ", port.Port, port.Protocol, port.Port)
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

		portMappings = append(portMappings, interceptsv1.PortMapping{
			ServicePort:   port.Port,
			WorkspacePort: workspacePort,
			Protocol:      port.Protocol,
		})
	}

	// Create ServiceIntercept resource
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      interceptName,
			Namespace: targetNamespace,
		},
		Spec: interceptsv1.ServiceInterceptSpec{
			WorkspaceRef: corev1.ObjectReference{
				Name:      workspace.Name,
				Namespace: workspace.Namespace,
			},
			ServiceRef: corev1.ObjectReference{
				Name:      service.Name,
				Namespace: targetNamespace,
			},
			PortMappings: portMappings,
		},
	}

	if err := WsClient.K8sClient.Create(ctx, intercept); err != nil {
		return fmt.Errorf("failed to create service intercept: %w", err)
	}

	fmt.Printf("\n[✓] Service intercept created successfully\n")
	fmt.Printf("Service '%s' is now being intercepted by workspace '%s'\n\n", service.Name, workspace.Name)
	fmt.Println("Port mappings:")
	for _, mapping := range portMappings {
		fmt.Printf("  %d (service) → %d (workspace)\n", mapping.ServicePort, mapping.WorkspacePort)
	}
	fmt.Printf("\nTraffic to %s.%s.svc.cluster.local will now be routed to your workspace.\n", service.Name, targetNamespace)

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

	// List all intercepts for this workspace
	fmt.Println("Loading active intercepts...")
	interceptList := &interceptsv1.ServiceInterceptList{}
	if err := WsClient.K8sClient.List(ctx, interceptList,
		client.InNamespace(targetNamespace),
		client.MatchingLabels{
			"intercepts.kloudlite.io/workspace-name": workspace.Name,
		},
	); err != nil {
		return fmt.Errorf("failed to list service intercepts: %w", err)
	}

	if len(interceptList.Items) == 0 {
		return fmt.Errorf("no active service intercepts found")
	}

	// Use embedded fzf to select intercept
	selectedIntercept, err := selectInterceptWithFzf(interceptList.Items)
	if err != nil {
		return err
	}

	// Delete the selected intercept
	if err := WsClient.K8sClient.Delete(ctx, selectedIntercept); err != nil {
		return fmt.Errorf("failed to delete service intercept: %w", err)
	}

	fmt.Printf("[✓] Service intercept stopped\n")
	fmt.Printf("Service '%s' is no longer being intercepted\n", selectedIntercept.Spec.ServiceRef.Name)
	fmt.Println("Normal traffic routing has been restored")

	return nil
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

	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace
	if targetNamespace == "" {
		return fmt.Errorf("workspace has no connected environment namespace")
	}

	// Find and delete the intercept
	interceptName := fmt.Sprintf("%s-%s", serviceName, workspace.Name)
	intercept := &interceptsv1.ServiceIntercept{}
	if err := WsClient.K8sClient.Get(ctx, types.NamespacedName{
		Name:      interceptName,
		Namespace: targetNamespace,
	}, intercept); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("no active intercept found for service '%s'", serviceName)
		}
		return fmt.Errorf("failed to get service intercept: %w", err)
	}

	if err := WsClient.K8sClient.Delete(ctx, intercept); err != nil {
		return fmt.Errorf("failed to delete service intercept: %w", err)
	}

	fmt.Printf("[✓] Service intercept stopped\n")
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

	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace
	if targetNamespace == "" {
		return fmt.Errorf("workspace has no connected environment namespace")
	}

	// List all intercepts in the target namespace for this workspace
	interceptList := &interceptsv1.ServiceInterceptList{}
	if err := WsClient.K8sClient.List(ctx, interceptList,
		client.InNamespace(targetNamespace),
		client.MatchingLabels{
			"intercepts.kloudlite.io/workspace-name": workspace.Name,
		},
	); err != nil {
		return fmt.Errorf("failed to list service intercepts: %w", err)
	}

	if len(interceptList.Items) == 0 {
		fmt.Println("No active service intercepts")
		fmt.Println("\nTo start intercepting a service, run:")
		fmt.Println("  kl intercept start")
		return nil
	}

	fmt.Printf("Active service intercepts (%d):\n\n", len(interceptList.Items))
	for _, intercept := range interceptList.Items {
		phase := intercept.Status.Phase
		if phase == "" {
			phase = "Creating"
		}

		fmt.Printf("  Service: %s\n", intercept.Spec.ServiceRef.Name)
		fmt.Printf("  Phase: %s\n", phase)
		fmt.Printf("  Port mappings:\n")
		for _, mapping := range intercept.Spec.PortMappings {
			fmt.Printf("    %d → %d (%s)\n", mapping.ServicePort, mapping.WorkspacePort, mapping.Protocol)
		}
		if intercept.Status.Message != "" {
			fmt.Printf("  Message: %s\n", intercept.Status.Message)
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

	targetNamespace := workspace.Status.ConnectedEnvironment.TargetNamespace
	if targetNamespace == "" {
		return fmt.Errorf("workspace has no connected environment namespace")
	}

	if serviceName != "" {
		// Show status for specific service
		interceptName := fmt.Sprintf("%s-%s", serviceName, workspace.Name)
		intercept := &interceptsv1.ServiceIntercept{}
		if err := WsClient.K8sClient.Get(ctx, types.NamespacedName{
			Name:      interceptName,
			Namespace: targetNamespace,
		}, intercept); err != nil {
			if apierrors.IsNotFound(err) {
				return fmt.Errorf("no active intercept found for service '%s'", serviceName)
			}
			return fmt.Errorf("failed to get service intercept: %w", err)
		}

		printInterceptStatus(intercept)
		return nil
	}

	// Show status for all intercepts
	interceptList := &interceptsv1.ServiceInterceptList{}
	if err := WsClient.K8sClient.List(ctx, interceptList,
		client.InNamespace(targetNamespace),
		client.MatchingLabels{
			"intercepts.kloudlite.io/workspace-name": workspace.Name,
		},
	); err != nil {
		return fmt.Errorf("failed to list service intercepts: %w", err)
	}

	if len(interceptList.Items) == 0 {
		fmt.Println("No active service intercepts")
		return nil
	}

	for i, intercept := range interceptList.Items {
		if i > 0 {
			fmt.Println("\n---")
		}
		printInterceptStatus(&intercept)
	}

	return nil
}

func printInterceptStatus(intercept *interceptsv1.ServiceIntercept) {
	fmt.Printf("Service: %s\n", intercept.Spec.ServiceRef.Name)
	fmt.Printf("Workspace: %s\n", intercept.Spec.WorkspaceRef.Name)
	fmt.Printf("Namespace: %s\n", intercept.Namespace)
	fmt.Printf("Phase: %s\n", intercept.Status.Phase)
	if intercept.Status.Message != "" {
		fmt.Printf("Message: %s\n", intercept.Status.Message)
	}

	fmt.Println("\nPort Mappings:")
	for _, mapping := range intercept.Spec.PortMappings {
		fmt.Printf("  %d (service) → %d (workspace) [%s]\n",
			mapping.ServicePort, mapping.WorkspacePort, mapping.Protocol)
	}

	if intercept.Status.WorkspacePodIP != "" {
		fmt.Printf("\nWorkspace Pod: %s\n", intercept.Status.WorkspacePodName)
		fmt.Printf("Workspace IP: %s\n", intercept.Status.WorkspacePodIP)
	}

	if len(intercept.Status.AffectedPodNames) > 0 {
		fmt.Printf("\nAffected Pods (%d):\n", len(intercept.Status.AffectedPodNames))
		for _, podName := range intercept.Status.AffectedPodNames {
			fmt.Printf("  - %s\n", podName)
		}
	}

	if intercept.Status.InterceptStartTime != nil {
		fmt.Printf("\nStart Time: %s\n", intercept.Status.InterceptStartTime.Format("2006-01-02 15:04:05"))
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

func selectInterceptWithFzf(intercepts []interceptsv1.ServiceIntercept) (*interceptsv1.ServiceIntercept, error) {
	// Create input for fzf
	interceptMap := make(map[string]*interceptsv1.ServiceIntercept)
	var items []string

	for i := range intercepts {
		intercept := &intercepts[i]
		phase := intercept.Status.Phase
		if phase == "" {
			phase = "Creating"
		}

		portStr := ""
		if len(intercept.Spec.PortMappings) > 0 {
			ports := make([]string, len(intercept.Spec.PortMappings))
			for j, mapping := range intercept.Spec.PortMappings {
				ports[j] = fmt.Sprintf("%d→%d", mapping.ServicePort, mapping.WorkspacePort)
			}
			portStr = strings.Join(ports, ", ")
		}

		line := fmt.Sprintf("%s (%s) - %s", intercept.Spec.ServiceRef.Name, phase, portStr)
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
