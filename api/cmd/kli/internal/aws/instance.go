package aws

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
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

func LaunchInstance(ctx context.Context, cfg aws.Config, amiID, subnetID, sgID, vpcID, secretKey, bucketName, k3sToken string, installationKey string, enableProtection bool) (string, error) {
	ec2Client := ec2.NewFromConfig(cfg)
	instanceName := fmt.Sprintf("kl-%s-instance", installationKey)
	profileName := fmt.Sprintf("kl-%s-role", installationKey)
	region := cfg.Region

	// Generate JWT secret for api-server
	jwtSecret, err := GenerateJWTSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	// Create cloud-init script to install K3s on startup
	userData := fmt.Sprintf(`#!/bin/bash
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

# Fetch instance IPs from EC2 metadata service
echo "Fetching instance metadata..."
METADATA_TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600" 2>/dev/null)
PRIVATE_IP=$(curl -H "X-aws-ec2-metadata-token: $METADATA_TOKEN" http://169.254.169.254/latest/meta-data/local-ipv4 2>/dev/null || echo "")
PUBLIC_IP=$(curl -H "X-aws-ec2-metadata-token: $METADATA_TOKEN" http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null || echo "")

echo "Instance Private IP: $PRIVATE_IP"
echo "Instance Public IP: $PUBLIC_IP"

# K3s configuration from Go
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
curl -fsSL "https://github.com/kloudlite/kloudlite/releases/latest/download/kli-linux-amd64" -o /usr/local/bin/kli
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
# Use the same JWT_SECRET as backend for shared secret authentication
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
  CLOUD_PROVIDER: "aws"
  INSTALLATION_KEY: "%s"
  AWS_VPC_ID: "%s"
  AWS_SECURITY_GROUP_ID: "%s"
  AWS_REGION: "%s"
  AWS_AMI_ID: "%s"
  AWS_PRIVATE_IP: "$PRIVATE_IP"
  AWS_PUBLIC_IP: "$PUBLIC_IP"
  K3S_VERSION: "$K3S_VERSION"
  K3S_SERVER_URL: "$K3S_SERVER_URL"
EOF

# Install manifests using embedded CRDs
echo "Installing Kloudlite manifests..."
export AWS_REGION="%s"
export S3_BUCKET="%s"
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

# Create AWS-specific MachineTypes
echo "Creating AWS MachineTypes..."
cat <<'MACHINEEOF' | kubectl apply -f -
%s
MACHINEEOF

if [ $? -eq 0 ]; then
  echo "AWS MachineTypes created successfully"
else
  echo "ERROR: Failed to create AWS MachineTypes"
fi

# API Server deployment is handled by kli install-manifests
# Wait for API Server to be ready
echo "Waiting for API Server to be ready..."
kubectl wait --for=condition=ready pod -l app=api-server -n kloudlite --timeout=300s || true

# Wait for Frontend to be ready (deployed via kli install-manifests)
echo "Waiting for Frontend to be ready..."
kubectl wait --for=condition=ready pod -l app=frontend -n kloudlite --timeout=300s || true

echo "Getting service endpoints..."
kubectl get svc -n kloudlite

# Setup K3s SQLite backup to S3
echo "Setting up K3s backup CronJob..."

# Save backup ConfigMap to manifests folder
cat <<EOF > /var/lib/rancher/k3s/server/manifests/k3s-backup-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: k3s-backup-config
  namespace: kloudlite
data:
  S3_BUCKET: "%s"
  AWS_REGION: "%s"
EOF

# Save backup CronJob and RBAC to manifests folder
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
  schedule: "*/30 * * * *"  # Every 30 minutes
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
              image: ghcr.io/kloudlite/kloudlite/k3s-backup:latest
              command:
                - /bin/bash
                - -c
                - |
                  set -euo pipefail

                  echo "Starting K3s backup at $(date)"

                  # K3s database location
                  DB_PATH="/var/lib/rancher/k3s/server/db/state.db"
                  BACKUP_FILE="/tmp/k3s-backup-$(date +%%Y%%m%%d-%%H%%M%%S).db"
                  S3_KEY="backups/k3s-backup-$(date +%%Y%%m%%d-%%H%%M%%S).db"

                  # Check if database exists
                  if [ ! -f "$DB_PATH" ]; then
                    echo "ERROR: K3s database not found at $DB_PATH"
                    exit 1
                  fi

                  # Copy database (SQLite backup)
                  echo "Backing up database..."
                  cp "$DB_PATH" "$BACKUP_FILE"

                  # Compress backup
                  echo "Compressing backup..."
                  gzip "$BACKUP_FILE"
                  BACKUP_FILE="${BACKUP_FILE}.gz"

                  # Upload to S3
                  echo "Uploading to S3: s3://${S3_BUCKET}/${S3_KEY}.gz"
                  aws s3 cp "$BACKUP_FILE" "s3://${S3_BUCKET}/${S3_KEY}.gz" --region ${AWS_REGION}

                  # Cleanup local backup
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
`, "v1.31.1+k3s1", k3sToken, secretKey, jwtSecret, jwtSecret, installationKey, vpcID, sgID, region, amiID, region, bucketName, manifests.AWSMachineTypes, bucketName, region)

	// Base64 encode the user data
	userDataEncoded := base64.StdEncoding.EncodeToString([]byte(userData))

	result, err := ec2Client.RunInstances(ctx, &ec2.RunInstancesInput{
		ImageId:      aws.String(amiID),
		InstanceType: types.InstanceTypeT3Medium,
		MinCount:     aws.Int32(1),
		MaxCount:     aws.Int32(1),
		UserData:     aws.String(userDataEncoded),
		NetworkInterfaces: []types.InstanceNetworkInterfaceSpecification{
			{
				DeviceIndex:              aws.Int32(0),
				SubnetId:                 aws.String(subnetID),
				Groups:                   []string{sgID},
				AssociatePublicIpAddress: aws.Bool(true),
			},
		},
		BlockDeviceMappings: []types.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &types.EbsBlockDevice{
					VolumeSize:          aws.Int32(100),
					VolumeType:          types.VolumeTypeGp3,
					DeleteOnTermination: aws.Bool(true),
				},
			},
		},
		IamInstanceProfile: &types.IamInstanceProfileSpecification{
			Name: aws.String(profileName),
		},
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeInstance,
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String(instanceName)},
					{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
					{Key: aws.String("Project"), Value: aws.String("kloudlite")},
					{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
					{Key: aws.String("kloudlite.io/installation-id"), Value: aws.String(installationKey)},
				},
			},
			{
				ResourceType: types.ResourceTypeVolume,
				Tags: []types.Tag{
					{Key: aws.String("Name"), Value: aws.String(fmt.Sprintf("%s-volume", instanceName))},
					{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
					{Key: aws.String("Project"), Value: aws.String("kloudlite")},
					{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
					{Key: aws.String("kloudlite.io/installation-id"), Value: aws.String(installationKey)},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to run instance: %w", err)
	}

	instanceID := *result.Instances[0].InstanceId

	// Enable termination protection if requested
	if enableProtection {
		_, err = ec2Client.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
			InstanceId: aws.String(instanceID),
			DisableApiTermination: &types.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
		})
		if err != nil {
			return instanceID, fmt.Errorf("failed to enable termination protection: %w", err)
		}
	}

	return instanceID, nil
}

func WaitForInstance(ctx context.Context, cfg aws.Config, instanceID string) (string, string, error) {
	ec2Client := ec2.NewFromConfig(cfg)

	waiter := ec2.NewInstanceRunningWaiter(ec2Client)
	err := waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}, 5*time.Minute)
	if err != nil {
		return "", "", fmt.Errorf("failed waiting for instance to be running: %w", err)
	}

	// Get instance details
	result, err := ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to describe instance: %w", err)
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return "", "", fmt.Errorf("instance not found")
	}

	instance := result.Reservations[0].Instances[0]
	publicIP := ""
	privateIP := ""

	if instance.PublicIpAddress != nil {
		publicIP = *instance.PublicIpAddress
	}
	if instance.PrivateIpAddress != nil {
		privateIP = *instance.PrivateIpAddress
	}

	return publicIP, privateIP, nil
}
