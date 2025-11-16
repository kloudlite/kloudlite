package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	workspacesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"cfg", "c"},
	Short:   "Manage workspace configuration",
	Long:    `View and update workspace configuration settings including display name, git config, and environment variables.`,
	Example: `  # View all configuration
  kl config get
  kl c get

  # View specific key
  kl config get display-name
  kl c get git.user-email

  # Update configuration
  kl config set display-name "My Workspace"
  kl c set git.user-name "John Doe"
  kl c set env.NODE_ENV production`,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "View workspace configuration",
	Long: `Display workspace configuration. When called without arguments, shows all configuration.
When called with a key, shows only that specific value.`,
	Example: `  # View all configuration
  kl config get
  kl c get

  # View specific values
  kl config get display-name
  kl c get git.user-email
  kl c get env.NODE_ENV`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleConfigGet(args)
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Update workspace configuration",
	Long: `Update a workspace configuration value.

Supported keys:
  - display-name
  - description
  - auto-stop
  - startup-script
  - dotfiles-repo
  - git.user-name
  - git.user-email
  - git.default-branch
  - env.<VAR_NAME>`,
	Example: `  kl config set display-name "Development Workspace"
  kl c set description "Main development environment"
  kl c set auto-stop true
  kl c set git.user-name "John Doe"
  kl c set git.user-email "john@example.com"
  kl c set env.NODE_ENV production`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleConfigSet(args)
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}

func handleConfigGet(args []string) error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return err
	}

	// If a specific key is requested
	if len(args) > 0 {
		key := args[0]
		return getConfigValue(workspace, key)
	}

	// Otherwise, display all configuration
	fmt.Println("Workspace Configuration:")
	fmt.Printf("  Display Name: %s\n", workspace.Spec.DisplayName)
	fmt.Printf("  Description: %s\n", workspace.Spec.Description)
	fmt.Printf("  Owner: %s\n", workspace.Spec.OwnedBy)
	fmt.Printf("  Folder Name: %s\n", workspace.Spec.FolderName)
	fmt.Printf("  VS Code Version: %s\n", workspace.Spec.VSCodeVersion)

	if len(workspace.Spec.Tags) > 0 {
		fmt.Printf("  Tags: %s\n", strings.Join(workspace.Spec.Tags, ", "))
	}

	if workspace.Spec.Settings != nil {
		fmt.Println("\nSettings:")
		fmt.Printf("  Auto Stop: %v\n", workspace.Spec.Settings.AutoStop)
		if workspace.Spec.Settings.IdleTimeout > 0 {
			fmt.Printf("  Idle Timeout: %d minutes\n", workspace.Spec.Settings.IdleTimeout)
		}
		if workspace.Spec.Settings.MaxRuntime > 0 {
			fmt.Printf("  Max Runtime: %d minutes\n", workspace.Spec.Settings.MaxRuntime)
		}
		if workspace.Spec.Settings.StartupScript != "" {
			fmt.Printf("  Startup Script: %s\n", workspace.Spec.Settings.StartupScript)
		}
		if workspace.Spec.Settings.DotfilesRepo != "" {
			fmt.Printf("  Dotfiles Repo: %s\n", workspace.Spec.Settings.DotfilesRepo)
		}

		if workspace.Spec.Settings.GitConfig != nil {
			fmt.Println("\nGit Configuration:")
			fmt.Printf("  User Name: %s\n", workspace.Spec.Settings.GitConfig.UserName)
			fmt.Printf("  User Email: %s\n", workspace.Spec.Settings.GitConfig.UserEmail)
			fmt.Printf("  Default Branch: %s\n", workspace.Spec.Settings.GitConfig.DefaultBranch)
		}

		if len(workspace.Spec.Settings.EnvironmentVariables) > 0 {
			fmt.Println("\nEnvironment Variables:")
			for k, v := range workspace.Spec.Settings.EnvironmentVariables {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}

		if len(workspace.Spec.Settings.VSCodeExtensions) > 0 {
			fmt.Println("\nVS Code Extensions:")
			for _, ext := range workspace.Spec.Settings.VSCodeExtensions {
				fmt.Printf("  - %s\n", ext)
			}
		}
	}

	return nil
}

func getConfigValue(workspace *workspacesv1.Workspace, key string) error {
	switch key {
	case "display-name":
		fmt.Println(workspace.Spec.DisplayName)
	case "description":
		fmt.Println(workspace.Spec.Description)
	case "owner":
		fmt.Println(workspace.Spec.OwnedBy)
	case "folder-name":
		fmt.Println(workspace.Spec.FolderName)
	case "vscode-version":
		fmt.Println(workspace.Spec.VSCodeVersion)
	case "auto-stop":
		if workspace.Spec.Settings != nil {
			fmt.Println(workspace.Spec.Settings.AutoStop)
		}
	case "idle-timeout":
		if workspace.Spec.Settings != nil {
			fmt.Println(workspace.Spec.Settings.IdleTimeout)
		}
	case "max-runtime":
		if workspace.Spec.Settings != nil {
			fmt.Println(workspace.Spec.Settings.MaxRuntime)
		}
	case "git.user-name":
		if workspace.Spec.Settings != nil && workspace.Spec.Settings.GitConfig != nil {
			fmt.Println(workspace.Spec.Settings.GitConfig.UserName)
		}
	case "git.user-email":
		if workspace.Spec.Settings != nil && workspace.Spec.Settings.GitConfig != nil {
			fmt.Println(workspace.Spec.Settings.GitConfig.UserEmail)
		}
	case "git.default-branch":
		if workspace.Spec.Settings != nil && workspace.Spec.Settings.GitConfig != nil {
			fmt.Println(workspace.Spec.Settings.GitConfig.DefaultBranch)
		}
	default:
		// Check if it's an environment variable
		if strings.HasPrefix(key, "env.") {
			envKey := strings.TrimPrefix(key, "env.")
			if workspace.Spec.Settings != nil {
				if val, ok := workspace.Spec.Settings.EnvironmentVariables[envKey]; ok {
					fmt.Println(val)
					return nil
				}
			}
			return fmt.Errorf("environment variable %s not found", envKey)
		}
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func handleConfigSet(args []string) error {
	if err := InitClient(); err != nil {
		return err
	}

	key := args[0]
	value := args[1]

	ctx := context.Background()
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return err
	}

	// Ensure settings is initialized
	if workspace.Spec.Settings == nil {
		workspace.Spec.Settings = &workspacesv1.WorkspaceSettings{}
	}

	// Update the configuration based on key
	switch key {
	case "display-name":
		workspace.Spec.DisplayName = value
	case "description":
		workspace.Spec.Description = value
	case "auto-stop":
		workspace.Spec.Settings.AutoStop = value == "true"
	case "startup-script":
		workspace.Spec.Settings.StartupScript = value
	case "dotfiles-repo":
		workspace.Spec.Settings.DotfilesRepo = value
	case "git.user-name":
		if workspace.Spec.Settings.GitConfig == nil {
			workspace.Spec.Settings.GitConfig = &workspacesv1.GitConfig{}
		}
		workspace.Spec.Settings.GitConfig.UserName = value
	case "git.user-email":
		if workspace.Spec.Settings.GitConfig == nil {
			workspace.Spec.Settings.GitConfig = &workspacesv1.GitConfig{}
		}
		workspace.Spec.Settings.GitConfig.UserEmail = value
	case "git.default-branch":
		if workspace.Spec.Settings.GitConfig == nil {
			workspace.Spec.Settings.GitConfig = &workspacesv1.GitConfig{}
		}
		workspace.Spec.Settings.GitConfig.DefaultBranch = value
	default:
		// Check if it's an environment variable
		if strings.HasPrefix(key, "env.") {
			envKey := strings.TrimPrefix(key, "env.")
			if workspace.Spec.Settings.EnvironmentVariables == nil {
				workspace.Spec.Settings.EnvironmentVariables = make(map[string]string)
			}
			workspace.Spec.Settings.EnvironmentVariables[envKey] = value
		} else {
			return fmt.Errorf("unknown config key: %s", key)
		}
	}

	// Apply the update using a JSON patch to avoid conflicts
	patchData, err := json.Marshal(map[string]interface{}{
		"spec": workspace.Spec,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	patch := client.RawPatch(types.MergePatchType, patchData)
	if err := WsClient.Patch(ctx, workspace, patch); err != nil {
		return fmt.Errorf("failed to patch workspace: %w", err)
	}

	fmt.Printf("Configuration updated: %s = %s\n", key, value)

	return nil
}
