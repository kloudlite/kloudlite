package oci

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kli/internal"
	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/manifests"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
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

// GetInstanceName returns the instance name for an installation
func GetInstanceName(installationKey string) string {
	return fmt.Sprintf("kl-%s-instance", installationKey)
}

// LaunchInstance creates an OCI Compute Instance with K3s and Kloudlite components
func LaunchInstance(ctx context.Context, cfg *OCIConfig, imageID, subnetID, nsgID, secretKey, bucketName, k3sToken, installationKey, fullDomain, sshPublicKey string, enableDeletionProtection bool) (string, error) {
	computeClient, err := core.NewComputeClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create compute client: %w", err)
	}

	instanceName := GetInstanceName(installationKey)

	// Generate JWT secret for api-server
	jwtSecret, err := GenerateJWTSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	// Create startup script for K3s installation
	startupScript := generateStartupScript(cfg, secretKey, bucketName, k3sToken, installationKey, fullDomain, jwtSecret, subnetID, nsgID)

	// Check if instance already exists
	existingID, err := findInstanceByName(ctx, computeClient, cfg.CompartmentOCID, instanceName)
	if err == nil && existingID != "" {
		return existingID, nil
	}

	tags := freeformTags(installationKey)

	// Instance shape configuration: VM.Standard.E4.Flex with 1 OCPU, 8GB RAM
	ocpus := float32(1)
	memoryInGBs := float32(8)

	// Use cloud-init for the startup script
	metadata := map[string]string{
		"user_data": base64.StdEncoding.EncodeToString([]byte(startupScript)),
	}
	if sshPublicKey != "" {
		metadata["ssh_authorized_keys"] = sshPublicKey
	}

	// Look up the actual availability domain name (format is like "FwKk:AP-MUMBAI-1-AD-1")
	ad, err := getFirstAvailabilityDomain(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to find availability domain: %w", err)
	}

	shape := "VM.Standard.E4.Flex"
	bootVolumeSizeInGBs := int64(100)
	launchResp, err := computeClient.LaunchInstance(ctx, core.LaunchInstanceRequest{
		LaunchInstanceDetails: core.LaunchInstanceDetails{
			AvailabilityDomain: &ad,
			CompartmentId:      &cfg.CompartmentOCID,
			DisplayName:        &instanceName,
			Shape:              &shape,
			ShapeConfig: &core.LaunchInstanceShapeConfigDetails{
				Ocpus:       &ocpus,
				MemoryInGBs: &memoryInGBs,
			},
			SourceDetails: core.InstanceSourceViaImageDetails{
				ImageId:             &imageID,
				BootVolumeSizeInGBs: &bootVolumeSizeInGBs,
			},
			CreateVnicDetails: &core.CreateVnicDetails{
				SubnetId:       &subnetID,
				NsgIds:         []string{nsgID},
				AssignPublicIp: boolPtr(true),
			},
			Metadata: metadata,
			FreeformTags: tags,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			existingID, findErr := findInstanceByName(ctx, computeClient, cfg.CompartmentOCID, instanceName)
			if findErr == nil && existingID != "" {
				return existingID, nil
			}
		}
		return "", fmt.Errorf("failed to create instance: %w", err)
	}

	return *launchResp.Id, nil
}

// WaitForInstance waits for the instance to be running and returns its IPs
func WaitForInstance(ctx context.Context, cfg *OCIConfig, instanceID string) (string, string, error) {
	computeClient, err := core.NewComputeClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", "", fmt.Errorf("failed to create compute client: %w", err)
	}

	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", "", fmt.Errorf("failed to create virtual network client: %w", err)
	}

	// Poll for instance status
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := computeClient.GetInstance(ctx, core.GetInstanceRequest{
			InstanceId: &instanceID,
		})
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.LifecycleState == core.InstanceLifecycleStateRunning {
			// Get VNIC attachments to find IPs
			publicIP, privateIP := getInstanceIPs(ctx, computeClient, vnClient, cfg.CompartmentOCID, instanceID)
			return publicIP, privateIP, nil
		}

		if resp.LifecycleState == core.InstanceLifecycleStateTerminated ||
			resp.LifecycleState == core.InstanceLifecycleStateTerminating {
			return "", "", fmt.Errorf("instance entered terminated state")
		}

		time.Sleep(5 * time.Second)
	}

	return "", "", fmt.Errorf("timeout waiting for instance to be running")
}

// DeleteInstance terminates an OCI instance
func DeleteInstance(ctx context.Context, cfg *OCIConfig, instanceID string) error {
	computeClient, err := core.NewComputeClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}

	preserveBootVolume := false
	_, err = computeClient.TerminateInstance(ctx, core.TerminateInstanceRequest{
		InstanceId:         &instanceID,
		PreserveBootVolume: &preserveBootVolume,
	})
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to terminate instance: %w", err)
	}

	return nil
}

// FindInstanceByInstallationKey finds the main instance by installation key (display name)
func FindInstanceByInstallationKey(ctx context.Context, cfg *OCIConfig, installationKey string) (string, error) {
	computeClient, err := core.NewComputeClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create compute client: %w", err)
	}

	instanceName := GetInstanceName(installationKey)
	return findInstanceByName(ctx, computeClient, cfg.CompartmentOCID, instanceName)
}

// FindAllInstancesByTag finds ALL instances with the installation-id freeform tag,
// including workmachines and other sub-resources created by the platform.
// This is the OCI equivalent of AWS's tag:kloudlite.io/installation-id filter.
func FindAllInstancesByTag(ctx context.Context, cfg *OCIConfig, installationKey string) ([]string, error) {
	computeClient, err := core.NewComputeClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	var instanceIDs []string
	var page *string

	for {
		resp, err := computeClient.ListInstances(ctx, core.ListInstancesRequest{
			CompartmentId: &cfg.CompartmentOCID,
			Page:          page,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list instances: %w", err)
		}

		for _, inst := range resp.Items {
			if inst.Id == nil {
				continue
			}
			// Skip terminated/terminating instances
			if inst.LifecycleState == core.InstanceLifecycleStateTerminated ||
				inst.LifecycleState == core.InstanceLifecycleStateTerminating {
				continue
			}
			// Match by freeform tag: installation-id
			if tagVal, ok := inst.FreeformTags["installation-id"]; ok && tagVal == installationKey {
				instanceIDs = append(instanceIDs, *inst.Id)
			}
		}

		if resp.OpcNextPage == nil {
			break
		}
		page = resp.OpcNextPage
	}

	return instanceIDs, nil
}

// TerminateAllInstances terminates all instances in the list.
// For each instance it first tries to disable termination protection (if any).
func TerminateAllInstances(ctx context.Context, cfg *OCIConfig, instanceIDs []string) error {
	computeClient, err := core.NewComputeClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create compute client: %w", err)
	}

	preserveBootVolume := false

	for _, instanceID := range instanceIDs {
		_, err := computeClient.TerminateInstance(ctx, core.TerminateInstanceRequest{
			InstanceId:         &instanceID,
			PreserveBootVolume: &preserveBootVolume,
		})
		if err != nil {
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				continue
			}
			fmt.Printf("    Warning: Failed to terminate instance %s: %v\n", instanceID, err)
		}
	}

	return nil
}

// findInstanceByName finds an instance by display name in the compartment
func findInstanceByName(ctx context.Context, computeClient core.ComputeClient, compartmentID, name string) (string, error) {
	resp, err := computeClient.ListInstances(ctx, core.ListInstancesRequest{
		CompartmentId: &compartmentID,
		DisplayName:   &name,
	})
	if err != nil {
		return "", err
	}

	for _, inst := range resp.Items {
		if inst.Id != nil &&
			inst.LifecycleState != core.InstanceLifecycleStateTerminated &&
			inst.LifecycleState != core.InstanceLifecycleStateTerminating {
			return *inst.Id, nil
		}
	}

	return "", fmt.Errorf("instance %s not found", name)
}

// getInstanceIPs returns the public and private IPs of an instance
func getInstanceIPs(ctx context.Context, computeClient core.ComputeClient, vnClient core.VirtualNetworkClient, compartmentID, instanceID string) (string, string) {
	// List VNIC attachments
	resp, err := computeClient.ListVnicAttachments(ctx, core.ListVnicAttachmentsRequest{
		CompartmentId: &compartmentID,
		InstanceId:    &instanceID,
	})
	if err != nil {
		return "", ""
	}

	for _, va := range resp.Items {
		if va.VnicId == nil || va.LifecycleState != core.VnicAttachmentLifecycleStateAttached {
			continue
		}

		vnicResp, err := vnClient.GetVnic(ctx, core.GetVnicRequest{
			VnicId: va.VnicId,
		})
		if err != nil {
			continue
		}

		publicIP := ""
		privateIP := ""
		if vnicResp.PublicIp != nil {
			publicIP = *vnicResp.PublicIp
		}
		if vnicResp.PrivateIp != nil {
			privateIP = *vnicResp.PrivateIp
		}
		return publicIP, privateIP
	}

	return "", ""
}

// getFirstAvailabilityDomain returns the first available AD in the compartment
func getFirstAvailabilityDomain(ctx context.Context, cfg *OCIConfig) (string, error) {
	identityClient, err := identity.NewIdentityClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create identity client: %w", err)
	}

	resp, err := identityClient.ListAvailabilityDomains(ctx, identity.ListAvailabilityDomainsRequest{
		CompartmentId: &cfg.CompartmentOCID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list availability domains: %w", err)
	}

	for _, ad := range resp.Items {
		if ad.Name != nil {
			return *ad.Name, nil
		}
	}

	return "", fmt.Errorf("no availability domains found in compartment %s", cfg.CompartmentOCID)
}

// generateStartupScript creates the startup script for K3s installation
func generateStartupScript(cfg *OCIConfig, secretKey, bucketName, k3sToken, installationKey, fullDomain, jwtSecret, subnetID, nsgID string) string {
	return fmt.Sprintf(`#!/bin/bash
set -euo pipefail

# Log output to file
exec > >(tee -a /var/log/kloudlite-init.log)
exec 2>&1

echo "Starting Kloudlite installation at $(date)"

# Wait for any existing apt processes to finish (cloud-init apt setup)
echo "Waiting for apt locks to be released..."
while fuser /var/lib/apt/lists/lock /var/lib/dpkg/lock /var/lib/dpkg/lock-frontend >/dev/null 2>&1; do
  echo "  apt is locked, waiting..."
  sleep 5
done
echo "apt locks released"

# Update system
apt-get update -y
apt-get upgrade -y

# Install required packages
apt-get install -y curl wget git

# Fetch instance IPs from OCI Instance Metadata Service v2
echo "Fetching instance metadata..."
PRIVATE_IP=$(curl -H "Authorization: Bearer Oracle" http://169.254.169.254/opc/v2/vnics/ 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin)[0]['privateIp'])" 2>/dev/null || echo "")
PUBLIC_IP=$(curl -H "Authorization: Bearer Oracle" http://169.254.169.254/opc/v2/vnics/ 2>/dev/null | python3 -c "import sys,json; v=json.load(sys.stdin)[0]; print(v.get('publicIp',''))" 2>/dev/null || echo "")

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
  CLOUD_PROVIDER: "oci"
  INSTALLATION_KEY: "%s"
  OCI_TENANCY: "%s"
  OCI_REGION: "%s"
  OCI_COMPARTMENT: "%s"
  OCI_SUBNET_ID: "%s"
  OCI_NSG_ID: "%s"
  OCI_PRIVATE_IP: "$PRIVATE_IP"
  OCI_PUBLIC_IP: "$PUBLIC_IP"
  K3S_VERSION: "$K3S_VERSION"
  K3S_SERVER_URL: "$K3S_SERVER_URL"
  HOSTED_SUBDOMAIN: "%s"
  REGISTRY_SERVICE_NAME: "cr.%s"
EOF

# Install manifests using embedded CRDs
echo "Installing Kloudlite manifests..."
export OCI_TENANCY="%s"
export OCI_REGION="%s"
export OCI_BUCKET="%s"
export AUTH_COOKIE_DOMAIN="%s"
export CLOUDFLARE_DNS_DOMAIN="khost.dev"
export KLOUDLITE_REGISTRY_HOST="cr.%s"
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

# Create OCI-specific MachineTypes
echo "Creating OCI MachineTypes..."
cat <<'MACHINEEOF' | kubectl apply -f -
%s
MACHINEEOF

if [ $? -eq 0 ]; then
  echo "OCI MachineTypes created successfully"
else
  echo "ERROR: Failed to create OCI MachineTypes"
fi

# Wait for API Server to be ready
echo "Waiting for API Server to be ready..."
kubectl wait --for=condition=ready pod -l app=api-server -n kloudlite --timeout=300s || true

# Wait for Frontend to be ready
echo "Waiting for Frontend to be ready..."
kubectl wait --for=condition=ready pod -l app=frontend -n kloudlite --timeout=300s || true

echo "Getting service endpoints..."
kubectl get svc -n kloudlite

# Setup K3s backup to OCI Object Storage
echo "Setting up K3s backup CronJob..."

# Save backup ConfigMap
cat <<EOF > /var/lib/rancher/k3s/server/manifests/k3s-backup-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k3s-backup-config
  namespace: kloudlite
data:
  OCI_BUCKET: "%s"
  OCI_NAMESPACE: "%s"
  OCI_REGION: "%s"
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
              image: ghcr.io/oracle/oci-cli:latest
              command:
                - /bin/bash
                - -c
                - |
                  set -euo pipefail

                  echo "Starting K3s backup at $(date)"

                  DB_PATH="/var/lib/rancher/k3s/server/db/state.db"
                  BACKUP_FILE="/tmp/k3s-backup-$(date +%%Y%%m%%d-%%H%%M%%S).db"
                  OCI_KEY="backups/k3s-backup-$(date +%%Y%%m%%d-%%H%%M%%S).db"

                  if [ ! -f "$DB_PATH" ]; then
                    echo "ERROR: K3s database not found at $DB_PATH"
                    exit 1
                  fi

                  echo "Backing up database..."
                  cp "$DB_PATH" "$BACKUP_FILE"

                  echo "Compressing backup..."
                  gzip "$BACKUP_FILE"
                  BACKUP_FILE="${BACKUP_FILE}.gz"

                  echo "Uploading to OCI Object Storage..."
                  oci os object put --namespace "$OCI_NAMESPACE" --bucket-name "$OCI_BUCKET" --name "${OCI_KEY}.gz" --file "$BACKUP_FILE" --auth instance_principal

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
		installationKey, cfg.TenancyOCID, cfg.Region, cfg.CompartmentOCID, subnetID, nsgID, fullDomain, fullDomain,
		cfg.TenancyOCID, cfg.Region, bucketName, fullDomain, fullDomain,
		manifests.OCIMachineTypes,
		bucketName, cfg.TenancyOCID, cfg.Region)
}
