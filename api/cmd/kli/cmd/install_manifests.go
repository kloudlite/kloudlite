package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/certs"
	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/manifests"
	"github.com/spf13/cobra"
)

var installManifestsCmd = &cobra.Command{
	Use:   "install-manifests",
	Short: "Write embedded Kloudlite manifests to K3s manifests directory",
	Long:  `Writes the embedded CRDs, RBAC, API Server, Webhooks, and Frontend manifests to the K3s server manifests directory for auto-application.`,
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

		// Write kloudlite-ingress namespace
		ingressNsPath := filepath.Join(manifestsDir, "kloudlite-ingress-namespace.yaml")
		if err := os.WriteFile(ingressNsPath, []byte(manifests.KloudliteIngressNamespace), 0644); err != nil {
			return fmt.Errorf("failed to write kloudlite-ingress namespace: %w", err)
		}
		fmt.Printf("✓ Written kloudlite-ingress namespace to %s\n", ingressNsPath)

		// Write kloudlite-hostmanager namespace
		hostmanagerNsPath := filepath.Join(manifestsDir, "kloudlite-hostmanager-namespace.yaml")
		if err := os.WriteFile(hostmanagerNsPath, []byte(manifests.KloudliteHostmanagerNamespace), 0644); err != nil {
			return fmt.Errorf("failed to write kloudlite-hostmanager namespace: %w", err)
		}
		fmt.Printf("✓ Written kloudlite-hostmanager namespace to %s\n", hostmanagerNsPath)

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

		// Generate webhook certificates
		fmt.Println("Generating webhook TLS certificates...")
		webhookCerts, err := certs.GenerateWebhookCertificates("api-server", "kloudlite")
		if err != nil {
			return fmt.Errorf("failed to generate webhook certificates: %w", err)
		}
		fmt.Println("✓ Generated webhook TLS certificates")

		// Create webhook TLS secret manifest (base64 encode the PEM data)
		webhookSecretManifest := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: webhook-server-cert
  namespace: kloudlite
type: kubernetes.io/tls
data:
  tls.crt: %s
  tls.key: %s
  ca.crt: %s
`,
			base64.StdEncoding.EncodeToString(webhookCerts.ServerCert),
			base64.StdEncoding.EncodeToString(webhookCerts.ServerKey),
			base64.StdEncoding.EncodeToString(webhookCerts.CACert))

		webhookSecretPath := filepath.Join(manifestsDir, "webhook-tls-secret.yaml")
		if err := os.WriteFile(webhookSecretPath, []byte(webhookSecretManifest), 0644); err != nil {
			return fmt.Errorf("failed to write webhook TLS secret: %w", err)
		}
		fmt.Printf("✓ Written webhook TLS secret to %s\n", webhookSecretPath)

		// Update webhooks manifest with CA bundle
		webhooksManifest := strings.ReplaceAll(manifests.Webhooks, `caBundle: ""`, fmt.Sprintf(`caBundle: %s`, webhookCerts.CABundle))
		webhooksPath := filepath.Join(manifestsDir, "webhooks.yaml")
		if err := os.WriteFile(webhooksPath, []byte(webhooksManifest), 0644); err != nil {
			return fmt.Errorf("failed to write Webhooks: %w", err)
		}
		fmt.Printf("✓ Written Webhooks to %s\n", webhooksPath)

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
