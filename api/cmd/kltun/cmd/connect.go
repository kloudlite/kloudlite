package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/api"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/daemon"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/spinner"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/truststore"
	"github.com/spf13/cobra"
)

var (
	connectToken  string
	connectServer string
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to Kloudlite VPN",
	Long: `Connect to Kloudlite VPN using your authentication token.

The connection runs in the background via the kltun daemon. Once connected,
you can access your Kloudlite workspace and services.

IMPORTANT: For security reasons, credentials are NOT saved to disk. You must
provide --token and --server flags on every connection.`,
	Example: `  # Connect with token and server (required every time)
  kltun connect --token YOUR_TOKEN --server https://subdomain.khost.dev`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConnect()
	},
}

func init() {
	connectCmd.Flags().StringVar(&connectToken, "token", "", "Authentication token")
	connectCmd.Flags().StringVar(&connectServer, "server", "", "Server URL (e.g., https://subdomain.khost.dev)")

	RootCmd.AddCommand(connectCmd)
}

func runConnect() error {
	fmt.Println()

	// Step 1: Start daemon
	sp := spinner.New("Starting daemon...")
	sp.Start()

	sm, err := daemon.NewServiceManager()
	if err != nil {
		sp.Stop(false)
		return fmt.Errorf("failed to create service manager: %w", err)
	}

	if err := sm.EnsureRunning(); err != nil {
		sp.Stop(false)
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	client := daemon.NewClient(sm.GetSocketPath())

	// Wait for daemon to be ready (max 10 seconds)
	daemonReady := false
	for i := 0; i < 100; i++ {
		if client.IsRunning() {
			daemonReady = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !daemonReady {
		sp.Stop(false)
		return fmt.Errorf("daemon failed to start within 10 seconds")
	}
	sp.Stop(true)

	// Step 2: Establish VPN connection
	sp = spinner.New("Establishing VPN connection...")
	sp.Start()

	result, err := client.VPNConnect(connectToken, connectServer)
	if err != nil {
		sp.Stop(false)
		return fmt.Errorf("failed to connect: %w", err)
	}
	sp.StopWithMessage(true, fmt.Sprintf("Connected (Session: %s)", result.SessionID))

	// Step 3: Install CA certificate
	if result.TunnelEndpoint != "" && result.PermanentToken != "" {
		sp = spinner.New("Installing CA certificate...")
		sp.Start()

		caCertInstalled := false
		var caCertError string

		// Create tunnel client to fetch CA cert
		tunnelClient := api.NewTunnelClient(result.TunnelEndpoint, result.PermanentToken)

		// Fetch CA cert with retry
		var caCert string
		var fetchErr error
		for attempt := 1; attempt <= 3; attempt++ {
			caCert, fetchErr = tunnelClient.GetCACert()
			if fetchErr == nil && caCert != "" {
				break
			}
			if attempt < 3 {
				time.Sleep(time.Duration(attempt) * time.Second)
			}
		}

		if fetchErr != nil {
			caCertError = fmt.Sprintf("failed to fetch: %v", fetchErr)
		} else if caCert == "" {
			caCertError = "CA certificate is empty"
		} else {
			// Write CA cert to temp file
			certFile := filepath.Join(os.TempDir(), fmt.Sprintf("kltun-ca-%s.crt", result.SessionID))
			if err := os.WriteFile(certFile, []byte(caCert), 0o600); err != nil {
				caCertError = fmt.Sprintf("failed to write: %v", err)
			} else {
				defer os.Remove(certFile)
				stores := []string{"system", "nss", "java"}
				if err := truststore.InstallAll(certFile, stores); err != nil {
					caCertError = fmt.Sprintf("failed to install: %v", err)
				} else {
					caCertInstalled = true
				}
			}
		}

		if caCertInstalled {
			sp.Stop(true)
		} else {
			sp.StopWithMessage(false, fmt.Sprintf("CA certificate (%s)", caCertError))
		}
	}

	fmt.Println()
	fmt.Println("VPN connection is running in the background.")
	fmt.Println("Use 'kltun quit' to disconnect.")
	fmt.Println()

	return nil
}
