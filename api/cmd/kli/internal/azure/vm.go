package azure

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4"
	"github.com/kloudlite/kloudlite/api/cmd/kli/internal"
	"github.com/kloudlite/kloudlite/api/cmd/kli/internal/manifests"
	"golang.org/x/crypto/ssh"
)

// Default VM size (equivalent to AWS t3.medium: 2 vCPU, 8GB RAM)
const DefaultVMSize = "Standard_B2ms"

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

// SSHKeyPair contains a generated SSH key pair
type SSHKeyPair struct {
	PublicKey  string // OpenSSH format public key
	PrivateKey string // PEM format private key
}

// GenerateSSHKeyPair generates an ED25519 SSH key pair
func GenerateSSHKeyPair() (*SSHKeyPair, error) {
	// Generate ED25519 key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ED25519 key: %w", err)
	}

	// Convert to SSH public key format
	sshPubKey, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH public key: %w", err)
	}
	pubKeyStr := string(ssh.MarshalAuthorizedKey(sshPubKey))

	// Convert private key to PEM format
	privKeyPEM, err := ssh.MarshalPrivateKey(privKey, "")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}
	privKeyStr := string(pem.EncodeToMemory(privKeyPEM))

	return &SSHKeyPair{
		PublicKey:  pubKeyStr,
		PrivateKey: privKeyStr,
	}, nil
}

// CreatePublicIP creates a public IP address for the VM
func CreatePublicIP(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	client, err := armnetwork.NewPublicIPAddressesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create public IP client: %w", err)
	}

	pipName := fmt.Sprintf("kl-%s-pip", installationKey)

	// Check if public IP already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, pipName, nil)
	if err == nil {
		return *existing.ID, nil
	}

	tags := map[string]*string{
		"Name":                      &pipName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation"),
		"kloudlite-installation-id": &installationKey,
	}

	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, pipName, armnetwork.PublicIPAddress{
		Location: &cfg.Location,
		Tags:     tags,
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: toIPAllocationMethod(armnetwork.IPAllocationMethodStatic),
			PublicIPAddressVersion:   toIPVersion(armnetwork.IPVersionIPv4),
		},
		SKU: &armnetwork.PublicIPAddressSKU{
			Name: toPublicIPSKUName(armnetwork.PublicIPAddressSKUNameStandard),
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create public IP: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to wait for public IP creation: %w", err)
	}

	return *result.ID, nil
}

// CreateNetworkInterface creates a network interface for the VM
func CreateNetworkInterface(ctx context.Context, cfg *AzureConfig, subnetID, nsgID, publicIPID, installationKey string) (string, error) {
	client, err := armnetwork.NewInterfacesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create network interface client: %w", err)
	}

	nicName := fmt.Sprintf("kl-%s-nic", installationKey)

	// Check if NIC already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, nicName, nil)
	if err == nil {
		return *existing.ID, nil
	}

	tags := map[string]*string{
		"Name":                      &nicName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation"),
		"kloudlite-installation-id": &installationKey,
	}

	ipConfigName := fmt.Sprintf("kl-%s-ipconfig", installationKey)

	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, nicName, armnetwork.Interface{
		Location: &cfg.Location,
		Tags:     tags,
		Properties: &armnetwork.InterfacePropertiesFormat{
			IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
				{
					Name: &ipConfigName,
					Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
						Subnet: &armnetwork.Subnet{
							ID: &subnetID,
						},
						PublicIPAddress: &armnetwork.PublicIPAddress{
							ID: &publicIPID,
						},
						PrivateIPAllocationMethod: toIPAllocationMethod(armnetwork.IPAllocationMethodDynamic),
					},
				},
			},
			NetworkSecurityGroup: &armnetwork.SecurityGroup{
				ID: &nsgID,
			},
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create network interface: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to wait for network interface creation: %w", err)
	}

	return *result.ID, nil
}

// LaunchVM creates an Azure VM with cloud-init for K3s installation
func LaunchVM(ctx context.Context, cfg *AzureConfig, imageRef *UbuntuImageReference, nicID, managedIdentityID,
	secretKey, storageAccountName, k3sToken, installationKey, vmSize, sshPublicKey string, enableProtection bool, fullDomain string,
	subnetID, nsgID string) (string, error) {

	client, err := armcompute.NewVirtualMachinesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create VM client: %w", err)
	}

	vmName := fmt.Sprintf("kl-%s-vm", installationKey)

	// Check if VM already exists
	existing, err := client.Get(ctx, cfg.ResourceGroup, vmName, nil)
	if err == nil {
		return *existing.ID, nil
	}

	// Use default VM size if not specified
	if vmSize == "" {
		vmSize = DefaultVMSize
	}

	// Generate JWT secret for api-server
	jwtSecret, err := GenerateJWTSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	// Create cloud-init script
	userData := generateCloudInitScript(cfg, secretKey, jwtSecret, storageAccountName, k3sToken, installationKey, fullDomain, subnetID, nsgID)
	userDataEncoded := base64.StdEncoding.EncodeToString([]byte(userData))

	tags := map[string]*string{
		"Name":                      &vmName,
		"ManagedBy":                 strPtr("kloudlite"),
		"Project":                   strPtr("kloudlite"),
		"Purpose":                   strPtr("kloudlite-installation"),
		"kloudlite-installation-id": &installationKey,
	}

	adminUsername := "kloudlite"

	poller, err := client.BeginCreateOrUpdate(ctx, cfg.ResourceGroup, vmName, armcompute.VirtualMachine{
		Location: &cfg.Location,
		Tags:     tags,
		Identity: &armcompute.VirtualMachineIdentity{
			Type: toVMIdentityType(armcompute.ResourceIdentityTypeUserAssigned),
			UserAssignedIdentities: map[string]*armcompute.UserAssignedIdentitiesValue{
				managedIdentityID: {},
			},
		},
		Properties: &armcompute.VirtualMachineProperties{
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: toVMSize(armcompute.VirtualMachineSizeTypes(vmSize)),
			},
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: imageRef.ToImageReference(),
				OSDisk: &armcompute.OSDisk{
					Name:         strPtr(fmt.Sprintf("%s-osdisk", vmName)),
					CreateOption: toCreateOption(armcompute.DiskCreateOptionTypesFromImage),
					DiskSizeGB:   int32Ptr(100),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: toStorageAccountType(armcompute.StorageAccountTypesPremiumLRS),
					},
					DeleteOption: toDeleteOption(armcompute.DiskDeleteOptionTypesDelete),
				},
			},
			OSProfile: &armcompute.OSProfile{
				ComputerName:  &vmName,
				AdminUsername: &adminUsername,
				CustomData:    &userDataEncoded,
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: boolPtr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    strPtr(fmt.Sprintf("/home/%s/.ssh/authorized_keys", adminUsername)),
								KeyData: strPtr(sshPublicKey),
							},
						},
					},
					ProvisionVMAgent: boolPtr(true),
				},
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: &nicID,
						Properties: &armcompute.NetworkInterfaceReferenceProperties{
							Primary:      boolPtr(true),
							DeleteOption: toNetworkDeleteOption(armcompute.DeleteOptionsDelete),
						},
					},
				},
			},
			DiagnosticsProfile: &armcompute.DiagnosticsProfile{
				BootDiagnostics: &armcompute.BootDiagnostics{
					Enabled: boolPtr(true),
				},
			},
		},
	}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create VM: %w", err)
	}

	result, err := poller.PollUntilDone(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to wait for VM creation: %w", err)
	}

	return *result.ID, nil
}

// generateCloudInitScript creates the cloud-init script for K3s installation
func generateCloudInitScript(cfg *AzureConfig, secretKey, jwtSecret, storageAccountName, k3sToken, installationKey, fullDomain, subnetID, nsgID string) string {
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
apt-get install -y curl wget git jq

# Install Azure CLI for blob storage operations
curl -sL https://aka.ms/InstallAzureCLIDeb | bash

# Fetch instance IPs from Azure Instance Metadata Service (IMDS)
echo "Fetching instance metadata..."
PRIVATE_IP=$(curl -s -H Metadata:true "http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/privateIpAddress?api-version=2021-02-01&format=text")
PUBLIC_IP=$(curl -s -H Metadata:true "http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/publicIpAddress?api-version=2021-02-01&format=text" || echo "")

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

# Apply Secrets directly (secrets should not be in manifests folder)
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
  CLOUD_PROVIDER: "azure"
  INSTALLATION_KEY: "%s"
  AZURE_SUBSCRIPTION_ID: "%s"
  AZURE_RESOURCE_GROUP: "%s"
  AZURE_LOCATION: "%s"
  AZURE_SUBNET_ID: "%s"
  AZURE_NSG_ID: "%s"
  AZURE_PRIVATE_IP: "$PRIVATE_IP"
  AZURE_PUBLIC_IP: "$PUBLIC_IP"
  AZURE_STORAGE_ACCOUNT: "%s"
  K3S_VERSION: "$K3S_VERSION"
  K3S_SERVER_URL: "$K3S_SERVER_URL"
  HOSTED_SUBDOMAIN: "%s"
  REGISTRY_SERVICE_NAME: "cr.%s"
EOF

# Install manifests using embedded CRDs
echo "Installing Kloudlite manifests..."
export AZURE_STORAGE_ACCOUNT="%s"
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

# Create Azure-specific MachineTypes
echo "Creating Azure MachineTypes..."
cat <<'MACHINEEOF' | kubectl apply -f -
%s
MACHINEEOF

if [ $? -eq 0 ]; then
  echo "Azure MachineTypes created successfully"
else
  echo "ERROR: Failed to create Azure MachineTypes"
fi

# Wait for API Server to be ready
echo "Waiting for API Server to be ready..."
kubectl wait --for=condition=ready pod -l app=api-server -n kloudlite --timeout=300s || true

# Wait for Frontend to be ready
echo "Waiting for Frontend to be ready..."
kubectl wait --for=condition=ready pod -l app=frontend -n kloudlite --timeout=300s || true

echo "Getting service endpoints..."
kubectl get svc -n kloudlite

# Setup K3s backup to Azure Blob Storage
echo "Setting up K3s backup CronJob..."

# Save backup ConfigMap
cat <<EOF > /var/lib/rancher/k3s/server/manifests/k3s-backup-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k3s-backup-config
  namespace: kloudlite
data:
  AZURE_STORAGE_ACCOUNT: "%s"
  AZURE_CONTAINER_NAME: "k3s-backups"
EOF

# Save backup CronJob and RBAC
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
              image: mcr.microsoft.com/azure-cli:latest
              command:
                - /bin/bash
                - -c
                - |
                  set -euo pipefail

                  echo "Starting K3s backup at $(date)"

                  DB_PATH="/var/lib/rancher/k3s/server/db/state.db"
                  BACKUP_FILE="/tmp/k3s-backup-$(date +%%Y%%m%%d-%%H%%M%%S).db"
                  BLOB_NAME="backups/k3s-backup-$(date +%%Y%%m%%d-%%H%%M%%S).db"

                  if [ ! -f "$DB_PATH" ]; then
                    echo "ERROR: K3s database not found at $DB_PATH"
                    exit 1
                  fi

                  echo "Backing up database..."
                  cp "$DB_PATH" "$BACKUP_FILE"

                  echo "Compressing backup..."
                  gzip "$BACKUP_FILE"
                  BACKUP_FILE="${BACKUP_FILE}.gz"

                  echo "Uploading to Azure Blob Storage..."
                  az storage blob upload \
                    --account-name ${AZURE_STORAGE_ACCOUNT} \
                    --container-name ${AZURE_CONTAINER_NAME} \
                    --name "${BLOB_NAME}.gz" \
                    --file "$BACKUP_FILE" \
                    --auth-mode login

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
`, internal.K3sVersion, k3sToken, secretKey, jwtSecret, jwtSecret, installationKey, cfg.SubscriptionID, cfg.ResourceGroup, cfg.Location, subnetID, nsgID, storageAccountName, fullDomain, fullDomain, storageAccountName, fullDomain, fullDomain, manifests.AzureMachineTypes, storageAccountName)
}

// WaitForVM waits for the VM to be in running state and returns its IPs
func WaitForVM(ctx context.Context, cfg *AzureConfig, vmID string) (string, string, error) {
	client, err := armcompute.NewVirtualMachinesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create VM client: %w", err)
	}

	vmName := extractResourceName(vmID)

	// Wait for VM to be running
	timeout := 10 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		result, err := client.Get(ctx, cfg.ResourceGroup, vmName, &armcompute.VirtualMachinesClientGetOptions{
			Expand: toInstanceViewTypes(),
		})
		if err != nil {
			return "", "", fmt.Errorf("failed to get VM: %w", err)
		}

		if result.Properties != nil && result.Properties.InstanceView != nil {
			for _, status := range result.Properties.InstanceView.Statuses {
				if status.Code != nil && *status.Code == "PowerState/running" {
					// VM is running, get IPs
					return getVMIPs(ctx, cfg, vmName)
				}
			}
		}

		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		case <-time.After(10 * time.Second):
			// Continue polling
		}
	}

	return "", "", fmt.Errorf("VM did not become running within %v", timeout)
}

// getVMIPs retrieves the public and private IPs of a VM
func getVMIPs(ctx context.Context, cfg *AzureConfig, vmName string) (string, string, error) {
	nicClient, err := armnetwork.NewInterfacesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create NIC client: %w", err)
	}

	pipClient, err := armnetwork.NewPublicIPAddressesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create public IP client: %w", err)
	}

	nicName := fmt.Sprintf("kl-%s-nic", extractInstallationKey(vmName))

	nic, err := nicClient.Get(ctx, cfg.ResourceGroup, nicName, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to get NIC: %w", err)
	}

	var privateIP, publicIP string

	if nic.Properties != nil && len(nic.Properties.IPConfigurations) > 0 {
		ipConfig := nic.Properties.IPConfigurations[0]
		if ipConfig.Properties != nil {
			if ipConfig.Properties.PrivateIPAddress != nil {
				privateIP = *ipConfig.Properties.PrivateIPAddress
			}

			if ipConfig.Properties.PublicIPAddress != nil && ipConfig.Properties.PublicIPAddress.ID != nil {
				pipName := extractResourceName(*ipConfig.Properties.PublicIPAddress.ID)
				pip, err := pipClient.Get(ctx, cfg.ResourceGroup, pipName, nil)
				if err == nil && pip.Properties != nil && pip.Properties.IPAddress != nil {
					publicIP = *pip.Properties.IPAddress
				}
			}
		}
	}

	return publicIP, privateIP, nil
}

// TerminateVM deletes a VM and its associated resources
func TerminateVM(ctx context.Context, cfg *AzureConfig, installationKey string) error {
	vmClient, err := armcompute.NewVirtualMachinesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create VM client: %w", err)
	}

	vmName := fmt.Sprintf("kl-%s-vm", installationKey)

	// Check if VM exists
	_, err = vmClient.Get(ctx, cfg.ResourceGroup, vmName, nil)
	if err != nil {
		// VM doesn't exist
		return nil
	}

	// Delete VM (this will also delete NIC and OS disk due to DeleteOption settings)
	poller, err := vmClient.BeginDelete(ctx, cfg.ResourceGroup, vmName, nil)
	if err != nil {
		return fmt.Errorf("failed to start VM deletion: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete VM: %w", err)
	}

	return nil
}

// DeletePublicIP deletes a public IP address
func DeletePublicIP(ctx context.Context, cfg *AzureConfig, installationKey string) error {
	client, err := armnetwork.NewPublicIPAddressesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create public IP client: %w", err)
	}

	pipName := fmt.Sprintf("kl-%s-pip", installationKey)

	// Check if public IP exists
	_, err = client.Get(ctx, cfg.ResourceGroup, pipName, nil)
	if err != nil {
		// Public IP doesn't exist
		return nil
	}

	poller, err := client.BeginDelete(ctx, cfg.ResourceGroup, pipName, nil)
	if err != nil {
		return fmt.Errorf("failed to start public IP deletion: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete public IP: %w", err)
	}

	return nil
}

// DeleteNetworkInterface deletes a network interface
func DeleteNetworkInterface(ctx context.Context, cfg *AzureConfig, installationKey string) error {
	client, err := armnetwork.NewInterfacesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return fmt.Errorf("failed to create NIC client: %w", err)
	}

	nicName := fmt.Sprintf("kl-%s-nic", installationKey)

	// Check if NIC exists
	_, err = client.Get(ctx, cfg.ResourceGroup, nicName, nil)
	if err != nil {
		// NIC doesn't exist
		return nil
	}

	poller, err := client.BeginDelete(ctx, cfg.ResourceGroup, nicName, nil)
	if err != nil {
		return fmt.Errorf("failed to start NIC deletion: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to delete NIC: %w", err)
	}

	return nil
}

// FindVMByInstallationKey finds a VM by installation key
func FindVMByInstallationKey(ctx context.Context, cfg *AzureConfig, installationKey string) (string, error) {
	client, err := armcompute.NewVirtualMachinesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create VM client: %w", err)
	}

	vmName := fmt.Sprintf("kl-%s-vm", installationKey)

	result, err := client.Get(ctx, cfg.ResourceGroup, vmName, nil)
	if err != nil {
		return "", nil // Not found
	}

	return *result.ID, nil
}

// extractInstallationKey extracts the installation key from a resource name
func extractInstallationKey(resourceName string) string {
	// Resource names follow pattern: kl-{key}-{type}
	// Extract the key part
	if len(resourceName) > 3 && resourceName[:3] == "kl-" {
		// Find the last hyphen
		lastHyphen := -1
		for i := len(resourceName) - 1; i >= 0; i-- {
			if resourceName[i] == '-' {
				lastHyphen = i
				break
			}
		}
		if lastHyphen > 3 {
			return resourceName[3:lastHyphen]
		}
	}
	return resourceName
}

// Helper functions for creating typed pointers

func toIPAllocationMethod(m armnetwork.IPAllocationMethod) *armnetwork.IPAllocationMethod {
	return &m
}

func toIPVersion(v armnetwork.IPVersion) *armnetwork.IPVersion {
	return &v
}

func toPublicIPSKUName(n armnetwork.PublicIPAddressSKUName) *armnetwork.PublicIPAddressSKUName {
	return &n
}

func toVMIdentityType(t armcompute.ResourceIdentityType) *armcompute.ResourceIdentityType {
	return &t
}

func toVMSize(s armcompute.VirtualMachineSizeTypes) *armcompute.VirtualMachineSizeTypes {
	return &s
}

func toCreateOption(o armcompute.DiskCreateOptionTypes) *armcompute.DiskCreateOptionTypes {
	return &o
}

func toStorageAccountType(t armcompute.StorageAccountTypes) *armcompute.StorageAccountTypes {
	return &t
}

func toDeleteOption(o armcompute.DiskDeleteOptionTypes) *armcompute.DiskDeleteOptionTypes {
	return &o
}

func toNetworkDeleteOption(o armcompute.DeleteOptions) *armcompute.DeleteOptions {
	return &o
}

func toInstanceViewTypes() *armcompute.InstanceViewTypes {
	t := armcompute.InstanceViewTypesInstanceView
	return &t
}
