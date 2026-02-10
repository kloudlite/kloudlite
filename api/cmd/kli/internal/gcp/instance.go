package gcp

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kloudlite/kloudlite/api/cmd/kli/internal"
	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/manifests"
)

// GenerateK3sToken generates a random 64-character hexadecimal token for K3s agent authentication
func GenerateK3sToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 64 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateJWTSecret generates a random base64-encoded secret for JWT signing
func GenerateJWTSecret() (string, error) {
	bytes := make([]byte, 32) // 32 bytes for JWT secret
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// LaunchInstance creates a Compute Engine VM with K3s and Kloudlite components
func LaunchInstance(ctx context.Context, cfg *GCPConfig, imageURL, serviceAccountEmail, secretKey, bucketName, k3sToken, installationKey, fullDomain string, enableDeletionProtection bool) (string, error) {
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create instances client: %w", err)
	}
	defer instancesClient.Close()

	instanceName := fmt.Sprintf("kl-%s-instance", installationKey)
	networkTag := NetworkTag(installationKey)
	machineType := fmt.Sprintf("zones/%s/machineTypes/e2-medium", cfg.Zone)

	// Generate JWT secret for api-server
	jwtSecret, err := GenerateJWTSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	// Create startup script for K3s installation
	startupScript := generateStartupScript(cfg, secretKey, bucketName, k3sToken, installationKey, fullDomain, jwtSecret)

	// Check if instance already exists
	_, err = instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Project:  cfg.Project,
		Zone:     cfg.Zone,
		Instance: instanceName,
	})
	if err == nil {
		// Instance exists
		return instanceName, nil
	}

	instance := &computepb.Instance{
		Name:        ptrString(instanceName),
		MachineType: ptrString(machineType),
		Disks: []*computepb.AttachedDisk{
			{
				AutoDelete: ptrBool(true),
				Boot:       ptrBool(true),
				Type:       ptrString("PERSISTENT"),
				InitializeParams: &computepb.AttachedDiskInitializeParams{
					DiskSizeGb:  ptrInt64(100),
					DiskType:    ptrString(fmt.Sprintf("zones/%s/diskTypes/pd-balanced", cfg.Zone)),
					SourceImage: ptrString(imageURL),
				},
			},
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			{
				Network:    ptrString(GetNetworkURL(cfg.Project, "default")),
				Subnetwork: ptrString(GetSubnetworkURL(cfg.Project, cfg.Region, "default")),
				AccessConfigs: []*computepb.AccessConfig{
					{
						Name:        ptrString("External NAT"),
						Type:        ptrString("ONE_TO_ONE_NAT"),
						NetworkTier: ptrString("PREMIUM"),
					},
				},
			},
		},
		Tags: &computepb.Tags{
			Items: []string{networkTag},
		},
		ServiceAccounts: []*computepb.ServiceAccount{
			{
				Email:  ptrString(serviceAccountEmail),
				Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
			},
		},
		Metadata: &computepb.Metadata{
			Items: []*computepb.Items{
				{
					Key:   ptrString("startup-script"),
					Value: ptrString(startupScript),
				},
			},
		},
		Labels: map[string]string{
			"managed-by":      "kloudlite",
			"project":         "kloudlite",
			"purpose":         "kloudlite-installation",
			"installation-id": installationKey,
		},
		DeletionProtection: ptrBool(enableDeletionProtection),
	}

	op, err := instancesClient.Insert(ctx, &computepb.InsertInstanceRequest{
		Project:          cfg.Project,
		Zone:             cfg.Zone,
		InstanceResource: instance,
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return instanceName, nil
		}
		return "", fmt.Errorf("failed to create instance: %w", err)
	}

	// Wait for operation to complete
	if err := op.Wait(ctx); err != nil {
		return "", fmt.Errorf("failed waiting for instance creation: %w", err)
	}

	return instanceName, nil
}

// WaitForInstance waits for the instance to be running and returns its IPs
func WaitForInstance(ctx context.Context, cfg *GCPConfig, instanceName string) (string, string, error) {
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to create instances client: %w", err)
	}
	defer instancesClient.Close()

	// Poll for instance status
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		instance, err := instancesClient.Get(ctx, &computepb.GetInstanceRequest{
			Project:  cfg.Project,
			Zone:     cfg.Zone,
			Instance: instanceName,
		})
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if instance.Status != nil && *instance.Status == "RUNNING" {
			var publicIP, privateIP string

			if len(instance.NetworkInterfaces) > 0 {
				ni := instance.NetworkInterfaces[0]
				if ni.NetworkIP != nil {
					privateIP = *ni.NetworkIP
				}
				if len(ni.AccessConfigs) > 0 && ni.AccessConfigs[0].NatIP != nil {
					publicIP = *ni.AccessConfigs[0].NatIP
				}
			}

			return publicIP, privateIP, nil
		}

		time.Sleep(5 * time.Second)
	}

	return "", "", fmt.Errorf("timeout waiting for instance to be running")
}

// DeleteInstance terminates an instance
func DeleteInstance(ctx context.Context, cfg *GCPConfig, instanceName string) error {
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create instances client: %w", err)
	}
	defer instancesClient.Close()

	// Disable deletion protection first
	if err := DisableDeletionProtection(ctx, cfg, instanceName); err != nil {
		// Log but continue - instance might not have protection enabled
		fmt.Printf("Warning: Could not disable deletion protection: %v\n", err)
	}

	op, err := instancesClient.Delete(ctx, &computepb.DeleteInstanceRequest{
		Project:  cfg.Project,
		Zone:     cfg.Zone,
		Instance: instanceName,
	})
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	// Wait for operation to complete
	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for instance deletion: %w", err)
	}

	return nil
}

// DisableDeletionProtection disables deletion protection on an instance
func DisableDeletionProtection(ctx context.Context, cfg *GCPConfig, instanceName string) error {
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create instances client: %w", err)
	}
	defer instancesClient.Close()

	op, err := instancesClient.SetDeletionProtection(ctx, &computepb.SetDeletionProtectionInstanceRequest{
		Project:            cfg.Project,
		Zone:               cfg.Zone,
		Resource:           instanceName,
		DeletionProtection: ptrBool(false),
	})
	if err != nil {
		return fmt.Errorf("failed to disable deletion protection: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for deletion protection update: %w", err)
	}

	return nil
}

// FindInstanceByInstallationKey finds an instance by installation key
func FindInstanceByInstallationKey(ctx context.Context, cfg *GCPConfig, installationKey string) (string, error) {
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create instances client: %w", err)
	}
	defer instancesClient.Close()

	instanceName := fmt.Sprintf("kl-%s-instance", installationKey)

	instance, err := instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Project:  cfg.Project,
		Zone:     cfg.Zone,
		Instance: instanceName,
	})
	if err != nil {
		return "", fmt.Errorf("instance not found: %w", err)
	}

	return *instance.Name, nil
}

// GetInstanceName returns the instance name for an installation
func GetInstanceName(installationKey string) string {
	return fmt.Sprintf("kl-%s-instance", installationKey)
}

// generateStartupScript creates the startup script for K3s installation
func generateStartupScript(cfg *GCPConfig, secretKey, bucketName, k3sToken, installationKey, fullDomain, jwtSecret string) string {
	return fmt.Sprintf(`#!/bin/bash
set -euo pipefail

# Log output to file
exec > >(tee -a /var/log/kloudlite-init.log)
exec 2>&1

echo "Starting Kloudlite installation at $(date)"

# Update system
apt-get update -y
apt-get upgrade -y

# Install required packages
apt-get install -y curl wget git

# Fetch instance IPs from GCP metadata service
echo "Fetching instance metadata..."
PRIVATE_IP=$(curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/ip 2>/dev/null || echo "")
PUBLIC_IP=$(curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip 2>/dev/null || echo "")

echo "Instance Private IP: $PRIVATE_IP"
echo "Instance Public IP: $PUBLIC_IP"

# K3s configuration
K3S_VERSION="%s"
K3S_AGENT_TOKEN="%s"
K3S_SERVER_URL="https://$PRIVATE_IP:6443"

# Install K3s server with predefined token
echo "Installing K3s server..."
curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION="$K3S_VERSION" K3S_AGENT_TOKEN="$K3S_AGENT_TOKEN" sh -s - server \
  --disable traefik \
  --write-kubeconfig-mode 644

# Wait for K3s to be ready
echo "Waiting for K3s to be ready..."
until kubectl get nodes 2>/dev/null; do
  sleep 2
done

echo "K3s installation completed at $(date)"
echo "K3s Version: $K3S_VERSION"
echo "K3s Server URL: $K3S_SERVER_URL"

# Install Kloudlite components
echo "Installing Kloudlite API Server and Frontend..."

# Create namespace
kubectl create namespace kloudlite || true

# Create K3s manifests directory
mkdir -p /var/lib/rancher/k3s/server/manifests

# Download and install Kloudlite CLI
echo "Downloading Kloudlite CLI binary..."
KLI_RELEASE_TAG=$(curl -fsSL "https://api.github.com/repos/kloudlite/kloudlite/releases" | \
  grep -o '"tag_name": "kli-[^"]*"' | head -1 | cut -d'"' -f4)
if [ -z "$KLI_RELEASE_TAG" ]; then
  echo "ERROR: Could not find kli release"
  exit 1
fi
echo "Found kli release: $KLI_RELEASE_TAG"
curl -fsSL "https://github.com/kloudlite/kloudlite/releases/download/${KLI_RELEASE_TAG}/kli-linux-amd64" -o /usr/local/bin/kli
chmod +x /usr/local/bin/kli

K3S_AGENT_TOKEN=$(cat /var/lib/rancher/k3s/server/agent-token)

# Apply Secrets directly
echo "Creating API Server Secret..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: api-server-secret
  namespace: kloudlite
type: Opaque
stringData:
  INSTALLATION_SECRET: "%s"
  K3S_AGENT_TOKEN: "$K3S_AGENT_TOKEN"
  JWT_SECRET: "%s"
EOF

echo "Creating Frontend Secret..."
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: frontend-secrets
  namespace: kloudlite
type: Opaque
stringData:
  jwt-secret: "%s"
  installation-secret: "%s"
EOF

# Save ConfigMap to manifests folder for auto-apply
echo "Creating API Server ConfigMap..."
cat <<EOF > /var/lib/rancher/k3s/server/manifests/api-server-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-server-config
  namespace: kloudlite
data:
  PORT: "8080"
  CLOUD_PROVIDER: "gcp"
  INSTALLATION_KEY: "%s"
  GCP_PROJECT: "%s"
  GCP_REGION: "%s"
  GCP_ZONE: "%s"
  GCP_NETWORK: "default"
  GCP_SUBNETWORK: "default"
  GCP_PRIVATE_IP: "$PRIVATE_IP"
  GCP_PUBLIC_IP: "$PUBLIC_IP"
  K3S_VERSION: "$K3S_VERSION"
  K3S_SERVER_URL: "$K3S_SERVER_URL"
  HOSTED_SUBDOMAIN: "%s"
  REGISTRY_SERVICE_NAME: "cr.%s"
EOF

# Install manifests using embedded CRDs
echo "Installing Kloudlite manifests..."
export GCP_PROJECT="%s"
export GCP_REGION="%s"
export GCS_BUCKET="%s"
export AUTH_COOKIE_DOMAIN="%s"
export CLOUDFLARE_DNS_DOMAIN="khost.dev"
export KLOUDLITE_REGISTRY_HOST="cr.%s"
export INSTALLATION_TYPE="kloudlite-cloud"
if ! kli install-manifests; then
  echo "ERROR: Failed to install Kloudlite manifests"
  exit 1
fi

echo "CRDs and RBAC will be auto-applied by K3s"

# Wait for MachineType CRD to be registered
echo "Waiting for MachineType CRD to be registered..."
until kubectl get crd machinetypes.machines.kloudlite.io 2>/dev/null; do
  echo "  Waiting for CRD registration..."
  sleep 2
done
echo "MachineType CRD registered successfully"

# Create GCP-specific MachineTypes
echo "Creating GCP MachineTypes..."
cat <<'MACHINEEOF' | kubectl apply -f -
%s
MACHINEEOF

if [ $? -eq 0 ]; then
  echo "GCP MachineTypes created successfully"
else
  echo "ERROR: Failed to create GCP MachineTypes"
fi

# Wait for API Server to be ready
echo "Waiting for API Server to be ready..."
kubectl wait --for=condition=ready pod -l app=api-server -n kloudlite --timeout=300s || true

# Wait for Frontend to be ready
echo "Waiting for Frontend to be ready..."
kubectl wait --for=condition=ready pod -l app=frontend -n kloudlite --timeout=300s || true

echo "Getting service endpoints..."
kubectl get svc -n kloudlite

# Setup K3s backup to GCS
echo "Setting up K3s backup CronJob..."

# Save backup ConfigMap
cat <<EOF > /var/lib/rancher/k3s/server/manifests/k3s-backup-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k3s-backup-config
  namespace: kloudlite
data:
  GCS_BUCKET: "%s"
  GCP_PROJECT: "%s"
EOF

# Save backup CronJob
cat <<'BACKUP_EOF' > /var/lib/rancher/k3s/server/manifests/k3s-backup.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k3s-backup
  namespace: kloudlite
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k3s-backup
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k3s-backup
subjects:
  - kind: ServiceAccount
    name: k3s-backup
    namespace: kloudlite
roleRef:
  kind: ClusterRole
  name: k3s-backup
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: k3s-backup
  namespace: kloudlite
spec:
  schedule: "*/30 * * * *"
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            app: k3s-backup
        spec:
          serviceAccountName: k3s-backup
          hostNetwork: true
          hostPID: true
          restartPolicy: OnFailure
          containers:
            - name: backup
              image: google/cloud-sdk:slim
              command:
                - /bin/bash
                - -c
                - |
                  set -euo pipefail

                  echo "Starting K3s backup at $(date)"

                  DB_PATH="/var/lib/rancher/k3s/server/db/state.db"
                  BACKUP_FILE="/tmp/k3s-backup-$(date +%%Y%%m%%d-%%H%%M%%S).db"
                  GCS_KEY="backups/k3s-backup-$(date +%%Y%%m%%d-%%H%%M%%S).db"

                  if [ ! -f "$DB_PATH" ]; then
                    echo "ERROR: K3s database not found at $DB_PATH"
                    exit 1
                  fi

                  echo "Backing up database..."
                  cp "$DB_PATH" "$BACKUP_FILE"

                  echo "Compressing backup..."
                  gzip "$BACKUP_FILE"
                  BACKUP_FILE="${BACKUP_FILE}.gz"

                  echo "Uploading to GCS: gs://${GCS_BUCKET}/${GCS_KEY}.gz"
                  gsutil cp "$BACKUP_FILE" "gs://${GCS_BUCKET}/${GCS_KEY}.gz"

                  rm -f "$BACKUP_FILE"

                  echo "Backup completed successfully at $(date)"
              envFrom:
                - configMapRef:
                    name: k3s-backup-config
              volumeMounts:
                - name: k3s-data
                  mountPath: /var/lib/rancher/k3s
                  readOnly: true
              resources:
                requests:
                  memory: "128Mi"
                  cpu: "100m"
                limits:
                  memory: "256Mi"
                  cpu: "200m"
          volumes:
            - name: k3s-data
              hostPath:
                path: /var/lib/rancher/k3s
                type: Directory
BACKUP_EOF

echo "K3s backup manifests created successfully"

echo "Kloudlite installation completed successfully at $(date)!"
`, internal.K3sVersion, k3sToken, secretKey, jwtSecret, jwtSecret, secretKey,
		installationKey, cfg.Project, cfg.Region, cfg.Zone, fullDomain, fullDomain,
		cfg.Project, cfg.Region, bucketName, fullDomain, fullDomain,
		manifests.GCPMachineTypes,
		bucketName, cfg.Project)
}

func ptrBool(b bool) *bool {
	return &b
}

func ptrInt64(i int64) *int64 {
	return &i
}
