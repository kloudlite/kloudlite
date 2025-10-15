package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	fzf "github.com/junegunn/fzf/src"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var envCmd = &cobra.Command{
	Use:     "env",
	Aliases: []string{"e", "environment"},
	Short:   "Manage workspace environment connections",
	Long: `Manage workspace environment connections.

When connected to an environment, services in that environment can be accessed
using short DNS names instead of full qualified names.

Example: After connecting to 'production' environment, you can access
services like 'api-server' instead of 'api-server.env-production.svc.cluster.local'`,
	Example: `  # Show connection status
  kl env status
  kl e st

  # Connect to an environment
  kl env connect production
  kl e c production

  # Disconnect from environment
  kl env disconnect
  kl e d`,
}

var envConnectCmd = &cobra.Command{
	Use:     "connect [environment-name]",
	Aliases: []string{"c", "conn"},
	Short:   "Connect workspace to an environment",
	Long: `Connect workspace to an environment for simplified service access.

After connecting, you can access services in the environment using short names
instead of fully qualified domain names.

Example: If you connect to environment 'production' with target namespace
'env-production', you can access services like:
  - 'api-server' instead of 'api-server.env-production.svc.cluster.local'
  - 'database' instead of 'database.env-production.svc.cluster.local'

Note: The environment must be activated for the connection to work.

If no environment name is provided, an interactive list will be shown.`,
	Example: `  # Interactive environment selection
  kl env connect
  kl e c

  # Connect to specific environment
  kl env connect production
  kl e c production`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Interactive mode
			return handleEnvConnectInteractive()
		}
		// Direct mode
		return handleEnvConnect(args[0])
	},
}

var envDisconnectCmd = &cobra.Command{
	Use:     "disconnect",
	Aliases: []string{"d", "disc"},
	Short:   "Disconnect workspace from environment",
	Long:    `Disconnect workspace from the connected environment.`,
	Example: `  # Disconnect from environment
  kl env disconnect
  kl e d`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleEnvDisconnect()
	},
}

var envStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st", "stat"},
	Short:   "Show environment connection status",
	Long:    `Show the current environment connection status and available services.`,
	Example: `  # Show connection status
  kl env status
  kl e st`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleEnvStatus()
	},
}

func init() {
	// Add subcommands
	envCmd.AddCommand(envConnectCmd)
	envCmd.AddCommand(envDisconnectCmd)
	envCmd.AddCommand(envStatusCmd)

	// Register with root command
	RootCmd.AddCommand(envCmd)
}

func handleEnvConnectInteractive() error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// List all environments
	fmt.Println("Loading environments...")
	envList := &environmentsv1.EnvironmentList{}
	if err := WsClient.K8sClient.List(ctx, envList); err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	if len(envList.Items) == 0 {
		return fmt.Errorf("no environments found")
	}

	// Use native fzf for better control over layout
	selectedEnv, err := selectEnvironmentWithFzf(envList.Items)
	if err != nil {
		return err
	}

	return handleEnvConnect(selectedEnv.Name)
}

func handleEnvConnect(environmentName string) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get the environment to find its target namespace
	env := &environmentsv1.Environment{}
	if err := WsClient.K8sClient.Get(ctx, types.NamespacedName{
		Name:      environmentName,
		Namespace: WsClient.Namespace,
	}, env); err != nil {
		return fmt.Errorf("failed to get environment '%s': %w", environmentName, err)
	}

	if !env.Spec.Activated {
		fmt.Printf("⚠️  Warning: Environment '%s' is not activated\n", environmentName)
	}

	targetNamespace := env.Spec.TargetNamespace
	if targetNamespace == "" {
		return fmt.Errorf("environment '%s' has no target namespace configured", environmentName)
	}

	// Get the current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Update workspace spec with environment reference
	workspace.Spec.EnvironmentRef = &corev1.ObjectReference{
		Name:      environmentName,
		Namespace: WsClient.Namespace,
	}

	// Update the workspace spec
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Printf("[✓] Connected to environment '%s'\n", environmentName)
	fmt.Printf("Target Namespace: %s\n", targetNamespace)
	fmt.Println()
	fmt.Println("DNS search domain added. You can now access services using short names:")
	fmt.Printf("  Example: curl http://api-server:8080\n")
	fmt.Printf("  Instead of: curl http://api-server.%s.svc.cluster.local:8080\n", targetNamespace)

	return nil
}

func handleEnvDisconnect() error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get the current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Clear the environment reference
	workspace.Spec.EnvironmentRef = nil

	// Update the workspace spec
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	fmt.Println("[✓] Disconnected from environment")
	fmt.Println("DNS configuration will be updated by the controller")

	return nil
}

func handleEnvStatus() error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()

	// Get the current workspace
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get workspace: %w", err)
	}

	// Check if workspace has a connected environment
	if workspace.Status.ConnectedEnvironment == nil {
		fmt.Println("Status: Not connected to any environment")
		fmt.Println()
		fmt.Println("To connect to an environment, run:")
		fmt.Println("  kl env connect")
		return nil
	}

	fmt.Println("[✓] Connected to environment")
	fmt.Printf("Environment: %s\n", workspace.Status.ConnectedEnvironment.Name)
	fmt.Printf("Target Namespace: %s\n", workspace.Status.ConnectedEnvironment.TargetNamespace)
	fmt.Println()
	fmt.Printf("Services in '%s' can be accessed using short names\n", workspace.Status.ConnectedEnvironment.TargetNamespace)
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  curl http://api-server:8080")

	return nil
}

func selectEnvironmentWithFzf(envs []environmentsv1.Environment) (*environmentsv1.Environment, error) {
	// Create input for fzf
	envMap := make(map[string]*environmentsv1.Environment)
	var items []string

	for i := range envs {
		env := &envs[i]
		status := "inactive"
		if env.Spec.Activated {
			status = "active"
		}
		line := fmt.Sprintf("%s (%s) - %s", env.Name, status, env.Spec.TargetNamespace)
		items = append(items, line)
		envMap[line] = env
	}

	// Create temporary file for input
	tmpfile, err := os.CreateTemp("", "fzf-env-*.txt")
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
		"--prompt=Select environment: ",
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
		return nil, fmt.Errorf("environment selection cancelled")
	}

	if selected == "" {
		return nil, fmt.Errorf("no environment selected")
	}

	env, ok := envMap[selected]
	if !ok {
		return nil, fmt.Errorf("invalid environment selected")
	}

	return env, nil
}

