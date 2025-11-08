package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"st", "s"},
	Short:   "Show workspace status and information",
	Long:    `Display comprehensive workspace information including metadata, resource usage, access URLs, and timing information.`,
	Example: `  kl status
  kl st
  kl s`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return handleStatus()
	},
}

func handleStatus() error {
	if err := InitClient(); err != nil {
		return err
	}

	ctx := context.Background()
	workspace, err := WsClient.Get(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Workspace: %s\n", workspace.Spec.DisplayName)
	fmt.Printf("Name: %s\n", workspace.Name)
	fmt.Printf("Namespace: %s\n", workspace.Namespace)
	fmt.Printf("Owner: %s\n", workspace.Spec.OwnedBy)
	fmt.Printf("Phase: %s\n", workspace.Status.Phase)
	fmt.Printf("Status: %s\n", workspace.Spec.Status)

	if workspace.Status.Message != "" {
		fmt.Printf("Message: %s\n", workspace.Status.Message)
	}

	// Display resource usage
	if workspace.Status.ResourceUsage != nil {
		fmt.Println("\nResource Usage:")
		fmt.Printf("  CPU: %s\n", workspace.Status.ResourceUsage.CPU)
		fmt.Printf("  Memory: %s\n", workspace.Status.ResourceUsage.Memory)
		fmt.Printf("  Storage: %s\n", workspace.Status.ResourceUsage.Storage)
	}

	// Display resource quota
	if workspace.Spec.ResourceQuota != nil {
		fmt.Println("\nResource Quota:")
		if workspace.Spec.ResourceQuota.CPU != "" {
			fmt.Printf("  CPU: %s\n", workspace.Spec.ResourceQuota.CPU)
		}
		if workspace.Spec.ResourceQuota.Memory != "" {
			fmt.Printf("  Memory: %s\n", workspace.Spec.ResourceQuota.Memory)
		}
		if workspace.Spec.ResourceQuota.Storage != "" {
			fmt.Printf("  Storage: %s\n", workspace.Spec.ResourceQuota.Storage)
		}
		if workspace.Spec.ResourceQuota.GPUs > 0 {
			fmt.Printf("  GPUs: %d\n", workspace.Spec.ResourceQuota.GPUs)
		}
	}

	// Display access URLs
	if len(workspace.Status.AccessURLs) > 0 {
		fmt.Println("\nAccess URLs:")
		for service, url := range workspace.Status.AccessURLs {
			fmt.Printf("  %s: %s\n", service, url)
		}
	}

	// Display timing information
	if workspace.Status.StartTime != nil {
		fmt.Printf("\nStart Time: %s\n", workspace.Status.StartTime.Format(time.RFC3339))
	}
	if workspace.Status.LastActivityTime != nil {
		fmt.Printf("Last Activity: %s\n", workspace.Status.LastActivityTime.Format(time.RFC3339))
	}
	if workspace.Status.TotalRuntime > 0 {
		fmt.Printf("Total Runtime: %d minutes\n", workspace.Status.TotalRuntime)
	}

	// Display active connections
	if workspace.Status.ActiveConnections > 0 {
		fmt.Printf("Active Connections: %d\n", workspace.Status.ActiveConnections)
	}

	return nil
}
