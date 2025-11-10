package cmd

import (
	"fmt"
	"runtime"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/daemon"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the kltun daemon service",
	Long:  `Manage the kltun daemon service that runs in the background with elevated privileges.`,
}

var daemonInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the daemon service",
	Long:  `Install the daemon service to run in the background with elevated privileges.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		if err := sm.Install(); err != nil {
			return fmt.Errorf("failed to install daemon: %w", err)
		}

		return nil
	},
}

var daemonUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the daemon service",
	Long:  `Uninstall the daemon service and remove it from the system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		if err := sm.Uninstall(); err != nil {
			return fmt.Errorf("failed to uninstall daemon: %w", err)
		}

		return nil
	},
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon service",
	Long:  `Start the daemon service if it's not already running.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		if err := sm.Start(); err != nil {
			return fmt.Errorf("failed to start daemon: %w", err)
		}

		return nil
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon service",
	Long:  `Stop the daemon service if it's running.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		if err := sm.Stop(); err != nil {
			return fmt.Errorf("failed to stop daemon: %w", err)
		}

		return nil
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the daemon service status",
	Long:  `Check if the daemon service is installed and running.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		installed := sm.IsInstalled()
		running, err := sm.Status()
		if err != nil {
			return fmt.Errorf("failed to get daemon status: %w", err)
		}

		fmt.Printf("Platform: %s\n", runtime.GOOS)
		fmt.Printf("Installed: %v\n", installed)
		fmt.Printf("Running: %v\n", running)

		// Try to get daemon status via RPC
		if running {
			client := daemon.NewClient(sm.GetSocketPath())
			if status, err := client.Status(); err == nil {
				fmt.Printf("\nDaemon Status:\n")
				fmt.Printf("  Active Connections: %d\n", len(status.Connections))
				for _, conn := range status.Connections {
					fmt.Printf("    - Session: %s\n", conn.SessionID)
					fmt.Printf("      Server: %s\n", conn.Server)
					fmt.Printf("      Uptime: %d seconds\n", conn.Uptime)
				}
			}
		}

		return nil
	},
}

var daemonRunCmd = &cobra.Command{
	Use:    "run",
	Short:  "Run the daemon server (internal use)",
	Long:   `Run the daemon server. This command is used internally by the service manager.`,
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create and start server
		sm, err := daemon.NewServiceManager()
		if err != nil {
			return fmt.Errorf("failed to create service manager: %w", err)
		}

		server, err := daemon.NewServer()
		if err != nil {
			return fmt.Errorf("failed to create server: %w", err)
		}

		socketPath := sm.GetSocketPath()
		fmt.Printf("Starting daemon server on %s...\n", socketPath)

		if err := server.Start(socketPath); err != nil {
			return fmt.Errorf("server error: %w", err)
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonInstallCmd)
	daemonCmd.AddCommand(daemonUninstallCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonRunCmd)
}
