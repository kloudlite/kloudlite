package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
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
		webhookSecretPath := filepath.Join(manifestsDir, "webhook-tls-secret.yaml")
		webhooksPath := filepath.Join(manifestsDir, "webhooks.yaml")

		var webhookCerts *certs.WebhookCertificates
		// Check if webhook TLS secret already exists - don't regenerate
		if _, err := os.Stat(webhookSecretPath); err == nil {
			fmt.Printf("✓ Webhook TLS secret already exists at %s (skipping regeneration)\n", webhookSecretPath)
			// We still need the CA bundle for webhooks.yaml, but we can skip regeneration
			// The webhooks.yaml should also exist if the secret exists
			if _, err := os.Stat(webhooksPath); err == nil {
				fmt.Printf("✓ Webhooks manifest already exists at %s (skipping regeneration)\n", webhooksPath)
			}
		} else {
			fmt.Println("Generating webhook TLS certificates...")
			webhookCerts, err = certs.GenerateWebhookCertificates("api-server", "kloudlite")
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

			if err := os.WriteFile(webhookSecretPath, []byte(webhookSecretManifest), 0644); err != nil {
				return fmt.Errorf("failed to write webhook TLS secret: %w", err)
			}
			fmt.Printf("✓ Written webhook TLS secret to %s\n", webhookSecretPath)
		}

		// Generate wildcard TLS certificate for ingress proxy
		authCookieDomain := os.Getenv("AUTH_COOKIE_DOMAIN")
		if authCookieDomain != "" {
			wildcardSecretPath := filepath.Join(manifestsDir, "wildcard-tls-secret.yaml")

			// Check if wildcard TLS secret already exists - don't regenerate to avoid CA mismatch
			if _, err := os.Stat(wildcardSecretPath); err == nil {
				fmt.Printf("✓ Wildcard TLS secret already exists at %s (skipping regeneration)\n", wildcardSecretPath)
			} else {
				fmt.Println("Generating wildcard TLS certificate...")
				wildcardCerts, err := certs.GenerateWildcardCertificates(authCookieDomain)
				if err != nil {
					return fmt.Errorf("failed to generate wildcard certificates: %w", err)
				}
				fmt.Println("✓ Generated wildcard TLS certificate")

				// Create wildcard TLS secret manifest
				wildcardSecretManifest := fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: kloudlite-wildcard-cert-tls
  namespace: kloudlite
type: kubernetes.io/tls
data:
  tls.crt: %s
  tls.key: %s
  ca.crt: %s
`,
					base64.StdEncoding.EncodeToString(wildcardCerts.Cert),
					base64.StdEncoding.EncodeToString(wildcardCerts.Key),
					base64.StdEncoding.EncodeToString(wildcardCerts.CACert))

				if err := os.WriteFile(wildcardSecretPath, []byte(wildcardSecretManifest), 0644); err != nil {
					return fmt.Errorf("failed to write wildcard TLS secret: %w", err)
				}
				fmt.Printf("✓ Written wildcard TLS secret to %s\n", wildcardSecretPath)
			}
		} else {
			fmt.Println("⚠ Skipping wildcard TLS certificate (AUTH_COOKIE_DOMAIN not set)")
		}

		// Update webhooks manifest with CA bundle (only if we generated new certs)
		if webhookCerts != nil {
			webhooksManifest := strings.ReplaceAll(manifests.Webhooks, `caBundle: ""`, fmt.Sprintf(`caBundle: %s`, webhookCerts.CABundle))
			if err := os.WriteFile(webhooksPath, []byte(webhooksManifest), 0644); err != nil {
				return fmt.Errorf("failed to write Webhooks: %w", err)
			}
			fmt.Printf("✓ Written Webhooks to %s\n", webhooksPath)
		}

		// Write Frontend (substitute environment variables)
		frontendManifest := manifests.Frontend
		// AUTH_COOKIE_DOMAIN is the subdomain.baseDomain (e.g., beanbag.khost.dev)
		// CLOUDFLARE_DNS_DOMAIN is the base domain (e.g., khost.dev)
		cloudflareDNSDomain := os.Getenv("CLOUDFLARE_DNS_DOMAIN")
		if cloudflareDNSDomain == "" {
			cloudflareDNSDomain = "khost.dev"
		}
		frontendManifest = strings.ReplaceAll(frontendManifest, "${AUTH_COOKIE_DOMAIN}", authCookieDomain)
		frontendManifest = strings.ReplaceAll(frontendManifest, "${CLOUDFLARE_DNS_DOMAIN}", cloudflareDNSDomain)

		frontendPath := filepath.Join(manifestsDir, "frontend.yaml")
		if err := os.WriteFile(frontendPath, []byte(frontendManifest), 0644); err != nil {
			return fmt.Errorf("failed to write Frontend: %w", err)
		}
		fmt.Printf("✓ Written Frontend to %s\n", frontendPath)

		// Write Image Registry (substitute environment variables)
		imageRegistryManifest := manifests.ImageRegistry
		// Get region and bucket from environment or use defaults
		region := os.Getenv("AWS_REGION")
		if region == "" {
			// Try to get from EC2 metadata
			region = "us-east-1" // fallback
		}
		bucketName := os.Getenv("KLOUDLITE_S3_BUCKET")
		if bucketName == "" {
			bucketName = os.Getenv("S3_BUCKET")
		}

		// Only write image-registry if we have the required config
		registryHost := os.Getenv("KLOUDLITE_REGISTRY_HOST")
		if bucketName != "" && region != "" && registryHost != "" {
			imageRegistryManifest = strings.ReplaceAll(imageRegistryManifest, "${REGION}", region)
			imageRegistryManifest = strings.ReplaceAll(imageRegistryManifest, "${BUCKET_NAME}", bucketName)
			imageRegistryManifest = strings.ReplaceAll(imageRegistryManifest, "${REGISTRY_HOST}", registryHost)

			imageRegistryPath := filepath.Join(manifestsDir, "image-registry.yaml")
			if err := os.WriteFile(imageRegistryPath, []byte(imageRegistryManifest), 0644); err != nil {
				return fmt.Errorf("failed to write Image Registry: %w", err)
			}
			fmt.Printf("✓ Written Image Registry to %s\n", imageRegistryPath)
			// Note: Registry runs without authentication - access control is handled at ingress layer
		} else {
			fmt.Println("⚠ Skipping Image Registry (S3_BUCKET, AWS_REGION, or KLOUDLITE_REGISTRY_HOST not set)")
		}

		// Write Ingress Proxy (nginx with hostNetwork for exposing frontend)
		ingressProxyPath := filepath.Join(manifestsDir, "ingress-proxy.yaml")
		if err := os.WriteFile(ingressProxyPath, []byte(manifests.IngressProxy), 0644); err != nil {
			return fmt.Errorf("failed to write Ingress Proxy: %w", err)
		}
		fmt.Printf("✓ Written Ingress Proxy to %s\n", ingressProxyPath)

		// Write Local Path StorageClass (for PVC storage with pathPattern)
		localPathStorageClassPath := filepath.Join(manifestsDir, "local-path-storageclass.yaml")
		if err := os.WriteFile(localPathStorageClassPath, []byte(manifests.LocalPathStorageClass), 0644); err != nil {
			return fmt.Errorf("failed to write Local Path StorageClass: %w", err)
		}
		fmt.Printf("✓ Written Local Path StorageClass to %s\n", localPathStorageClassPath)

		// Write Local Path Provisioner Config (btrfs subvolume setup for snapshots)
		localPathConfigPath := filepath.Join(manifestsDir, "local-path-provisioner-config.yaml")
		if err := os.WriteFile(localPathConfigPath, []byte(manifests.LocalPathProvisionerConfig), 0644); err != nil {
			return fmt.Errorf("failed to write Local Path Provisioner Config: %w", err)
		}
		fmt.Printf("✓ Written Local Path Provisioner Config to %s\n", localPathConfigPath)

		fmt.Println("\nKloudlite manifests installed successfully!")
		fmt.Println("K3s will auto-apply these manifests on startup.")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(installManifestsCmd)
}

// runCommand executes a shell command and returns any error
func runCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
