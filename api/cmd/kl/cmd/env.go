package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
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

	// Use fuzzyfinder to select environment
	idx, err := fuzzyfinder.Find(
		envList.Items,
		func(i int) string {
			env := envList.Items[i]
			status := "inactive"
			if env.Spec.Activated {
				status = "active"
			}
			return fmt.Sprintf("%s (%s) - %s", env.Name, status, env.Spec.TargetNamespace)
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			env := envList.Items[i]
			preview := fmt.Sprintf("Name: %s\n", env.Name)
			preview += fmt.Sprintf("Target Namespace: %s\n", env.Spec.TargetNamespace)
			preview += fmt.Sprintf("Activated: %t\n", env.Spec.Activated)
			preview += fmt.Sprintf("Created By: %s\n", env.Spec.CreatedBy)
			if env.Status.State != "" {
				preview += fmt.Sprintf("State: %s\n", env.Status.State)
			}
			if env.Status.Message != "" {
				preview += fmt.Sprintf("Message: %s\n", env.Status.Message)
			}
			return preview
		}),
	)

	if err != nil {
		return fmt.Errorf("environment selection cancelled")
	}

	selectedEnv := envList.Items[idx]
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

	// Update /etc/resolv.conf with the new search domain
	if err := updateResolvConf(targetNamespace, true); err != nil {
		return fmt.Errorf("failed to update DNS configuration: %w", err)
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
	// Remove environment search domains from /etc/resolv.conf
	if err := updateResolvConf("", false); err != nil {
		return fmt.Errorf("failed to update DNS configuration: %w", err)
	}

	fmt.Println("[✓] Disconnected from environment")
	fmt.Println("Environment search domains removed from DNS configuration")

	return nil
}

func handleEnvStatus() error {
	// Read current DNS configuration from /etc/resolv.conf
	searchDomains, err := getSearchDomainsFromResolvConf()
	if err != nil {
		return fmt.Errorf("failed to read DNS configuration: %w", err)
	}

	// Find environment-specific search domains
	var envDomains []string
	for _, domain := range searchDomains {
		// Check if it's an environment namespace (e.g., env-xxx.svc.cluster.local)
		if strings.HasPrefix(domain, "env-") && strings.Contains(domain, ".svc.cluster.local") {
			envDomains = append(envDomains, domain)
		}
	}

	if len(envDomains) == 0 {
		fmt.Println("Status: Not connected to any environment")
		fmt.Println()
		fmt.Println("To connect to an environment, run:")
		fmt.Println("  kl env connect")
		return nil
	}

	fmt.Println("[✓] Connected to environment")
	fmt.Println()
	fmt.Println("Active search domains:")
	for _, domain := range envDomains {
		fmt.Printf("  - %s\n", domain)
		// Extract namespace from domain (e.g., env-sample from env-sample.svc.cluster.local)
		parts := strings.Split(domain, ".")
		if len(parts) > 0 {
			fmt.Printf("    → Services in '%s' can be accessed using short names\n", parts[0])
		}
	}
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("  curl http://api-server:8080")

	return nil
}

// updateResolvConf modifies /etc/resolv.conf to add or remove environment search domains
func updateResolvConf(targetNamespace string, add bool) error {
	resolvConfPath := "/etc/resolv.conf"

	// Read current resolv.conf
	file, err := os.Open(resolvConfPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", resolvConfPath, err)
	}
	defer file.Close()

	var lines []string
	var searchLine string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "search ") {
			searchLine = line
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read %s: %w", resolvConfPath, err)
	}

	// Parse existing search domains
	var searchDomains []string
	if searchLine != "" {
		parts := strings.Fields(searchLine)
		if len(parts) > 1 {
			searchDomains = parts[1:] // Skip "search" keyword
		}
	}

	// Remove any existing environment search domains
	var filteredDomains []string
	for _, domain := range searchDomains {
		// Keep non-environment domains
		if !strings.HasPrefix(domain, "env-") || !strings.Contains(domain, ".svc.cluster.local") {
			filteredDomains = append(filteredDomains, domain)
		}
	}

	// Add new environment domain if requested
	if add && targetNamespace != "" {
		newDomain := fmt.Sprintf("%s.svc.cluster.local", targetNamespace)
		// Add at the beginning for priority
		filteredDomains = append([]string{newDomain}, filteredDomains...)
	}

	// Build new search line
	if len(filteredDomains) > 0 {
		searchLine = "search " + strings.Join(filteredDomains, " ")
	} else {
		searchLine = ""
	}

	// Write updated resolv.conf directly (can't use temp file + rename for mounted files)
	// Open with O_WRONLY|O_TRUNC to truncate and write
	f, err := os.OpenFile(resolvConfPath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %w", resolvConfPath, err)
	}
	defer f.Close()

	writer := bufio.NewWriter(f)

	// Write search line first
	if searchLine != "" {
		if _, err := writer.WriteString(searchLine + "\n"); err != nil {
			return fmt.Errorf("failed to write search line: %w", err)
		}
	}

	// Write other lines
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}

// getSearchDomainsFromResolvConf reads search domains from /etc/resolv.conf
func getSearchDomainsFromResolvConf() ([]string, error) {
	file, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/resolv.conf: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "search ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				return parts[1:], nil // Return domains, skipping "search" keyword
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read /etc/resolv.conf: %w", err)
	}

	return []string{}, nil
}
