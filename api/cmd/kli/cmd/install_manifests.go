package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/manifests"
	"github.com/spf13/cobra"
)

var installManifestsCmd = &cobra.Command{
	Use:   "install-manifests",
	Short: "Write embedded Kloudlite manifests to K3s manifests directory",
	Long:  `Writes the embedded CRDs, RBAC, API Server, and Frontend manifests to the K3s server manifests directory for auto-application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manifestsDir := "/var/lib/rancher/k3s/server/manifests"

		// Create manifests directory if it doesn't exist
		if err := os.MkdirAll(manifestsDir, 0755); err != nil {
			return fmt.Errorf("failed to create manifests directory: %w", err)
		}

		// Write CRDs
		crdsPath := filepath.Join(manifestsDir, "kloudlite-crds.yaml")
		if err := os.WriteFile(crdsPath, []byte(manifests.CRDs), 0644); err != nil {
			return fmt.Errorf("failed to write CRDs: %w", err)
		}
		fmt.Printf("✓ Written CRDs to %s\n", crdsPath)

		// Write RBAC
		rbacPath := filepath.Join(manifestsDir, "api-server-rbac.yaml")
		if err := os.WriteFile(rbacPath, []byte(manifests.APIServerRBAC), 0644); err != nil {
			return fmt.Errorf("failed to write RBAC: %w", err)
		}
		fmt.Printf("✓ Written RBAC to %s\n", rbacPath)

		// Write API Server
		apiServerPath := filepath.Join(manifestsDir, "api-server.yaml")
		if err := os.WriteFile(apiServerPath, []byte(manifests.APIServer), 0644); err != nil {
			return fmt.Errorf("failed to write API Server: %w", err)
		}
		fmt.Printf("✓ Written API Server to %s\n", apiServerPath)

		// Write Frontend
		frontendPath := filepath.Join(manifestsDir, "frontend.yaml")
		if err := os.WriteFile(frontendPath, []byte(manifests.Frontend), 0644); err != nil {
			return fmt.Errorf("failed to write Frontend: %w", err)
		}
		fmt.Printf("✓ Written Frontend to %s\n", frontendPath)

		fmt.Println("\nKloudlite manifests installed successfully!")
		fmt.Println("K3s will auto-apply these manifests on startup.")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(installManifestsCmd)
}
