package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	fzf "github.com/junegunn/fzf/src"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	listOpts := []client.ListOption{
		client.InNamespace(WsClient.Namespace),
	}
	if err := WsClient.K8sClient.List(ctx, envList, listOpts...); err != nil {
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

	// Update workspace spec with environment connection
	workspace.Spec.EnvironmentConnection = &workspacev1.EnvironmentConnectionSpec{
		EnvironmentRef: corev1.ObjectReference{
			Name:      environmentName,
			Namespace: WsClient.Namespace,
		},
	}

	// Update the workspace spec
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	// Wait for the environment connection to sync
	// Use display name format: {owner}/{envName} to match what controller sets in status
	displayName := fmt.Sprintf("%s/%s", env.Spec.OwnedBy, env.Spec.Name)
	if err := waitForEnvironmentSync(displayName, targetNamespace, false); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	fmt.Println()
	fmt.Printf("[✓] Connected to environment '%s'\n", displayName)
	fmt.Println()
	fmt.Println("You can now access services in this environment using short names:")
	fmt.Println("  Example: curl http://api-server:8080")

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

	// Clear the environment connection (this also removes all intercepts)
	workspace.Spec.EnvironmentConnection = nil

	// Update the workspace spec
	if err := WsClient.Update(ctx, workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	// Wait for the disconnection to sync (pass empty strings for disconnect)
	if err := waitForEnvironmentSync("", "", true); err != nil {
		return fmt.Errorf("failed to disconnect: %w", err)
	}

	fmt.Println()
	fmt.Println("[✓] Disconnected from environment")

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

	fmt.Printf("[✓] Connected to environment '%s'\n", workspace.Status.ConnectedEnvironment.Name)
	fmt.Println()
	fmt.Println("Services in this environment can be accessed using short names:")
	fmt.Println("  Example: curl http://api-server:8080")

	// Read intercepts from context file (maintained by controller)
	fmt.Println()
	type ContextData struct {
		Environment string   `json:"environment"`
		Intercepts  []string `json:"intercepts"`
	}

	contextFile := "/tmp/kloudlite-context.json"
	contextBytes, err := os.ReadFile(contextFile)
	if err == nil {
		var contextData ContextData
		if err := json.Unmarshal(contextBytes, &contextData); err == nil && len(contextData.Intercepts) > 0 {
			fmt.Println("Active Service Intercepts:")
			for _, serviceName := range contextData.Intercepts {
				fmt.Printf("  - %s\n", serviceName)
			}
		} else {
			fmt.Println("No active service intercepts")
		}
	} else {
		fmt.Println("No active service intercepts")
	}

	return nil
}

// waitForEnvironmentSync waits for the workspace status to reflect the desired environment connection
func waitForEnvironmentSync(environmentName, targetNamespace string, isDisconnect bool) error {
	ctx := context.Background()
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Set up signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	if isDisconnect {
		fmt.Print("Disconnecting")
	} else {
		fmt.Print("Connecting")
	}

	for {
		select {
		case <-sigChan:
			fmt.Println(" interrupted!")
			return fmt.Errorf("interrupted by user")
		case <-timeout:
			fmt.Println(" timeout!")
			return fmt.Errorf("connection timed out after 30 seconds, please try again")
		case <-ticker.C:
			// Show progress
			fmt.Print(".")

			// Get current workspace status
			workspace, err := WsClient.Get(ctx)
			if err != nil {
				continue // Retry on transient errors
			}

			// Check if sync is complete
			if environmentName == "" {
				// Disconnecting - wait for ConnectedEnvironment to be nil
				if workspace.Status.ConnectedEnvironment == nil {
					fmt.Println(" done!")
					return nil
				}
			} else {
				// Connecting - wait for ConnectedEnvironment to match desired environment
				if workspace.Status.ConnectedEnvironment != nil &&
					workspace.Status.ConnectedEnvironment.Name == environmentName &&
					workspace.Status.ConnectedEnvironment.TargetNamespace == targetNamespace {
					fmt.Println(" done!")
					return nil
				}
			}
		}
	}
}

func selectEnvironmentWithFzf(envs []environmentsv1.Environment) (*environmentsv1.Environment, error) {
	// Create input for fzf
	envMap := make(map[string]*environmentsv1.Environment)
	var items []string

	for i := range envs {
		env := &envs[i]
		// Only show active environments (can't connect to inactive ones)
		if !env.Spec.Activated {
			continue
		}
		// Display format: {owner}/{envName}
		line := fmt.Sprintf("%s/%s", env.Spec.OwnedBy, env.Spec.Name)
		items = append(items, line)
		envMap[line] = env
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no active environments found")
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
