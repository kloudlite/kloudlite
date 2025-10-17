package aws

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/errors"
	workmachinev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
)

// GenerateK3sUserData generates the cloud-init user data script for K3s agent
// The script installs K3s agent and joins it to the existing cluster
func GenerateK3sUserData(wm *workmachinev1.WorkMachine, k3sToken string) (string, error) {
	if wm.Spec.AWSProvider == nil {
		return "", errors.NewInvalidConfigurationError("aws", "AWS provider configuration is required")
	}

	awsConfig := wm.Spec.AWSProvider

	if awsConfig.K3sServerURL == "" {
		return "", errors.NewInvalidConfigurationError("k3sServerURL", "K3s server URL is required")
	}

	if k3sToken == "" {
		return "", errors.NewInvalidConfigurationError("k3sToken", "K3s token is required")
	}

	// Build node labels
	nodeLabels := []string{
		fmt.Sprintf("kloudlite.io/workmachine=%s", wm.Name),
		fmt.Sprintf("kloudlite.io/owner=%s", wm.Spec.OwnedBy),
		"kloudlite.io/node-type=workmachine",
	}

	// Build node taints to isolate workspace pods
	nodeTaints := []string{
		fmt.Sprintf("kloudlite.io/workmachine=%s:NoSchedule", wm.Name),
	}

	userDataScript := fmt.Sprintf(`#!/bin/bash
set -euo pipefail

# Log all output to a file for debugging
exec > >(tee -a /var/log/k3s-install.log)
exec 2>&1

echo "Starting K3s agent installation for WorkMachine: %s"
echo "Timestamp: $(date)"

# Update system packages
echo "Updating system packages..."
apt-get update -y || yum update -y || true

# Install required dependencies
echo "Installing dependencies..."
apt-get install -y curl jq || yum install -y curl jq || true

# Get instance metadata
echo "Fetching EC2 instance metadata..."
INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
AVAILABILITY_ZONE=$(curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone)
PRIVATE_IP=$(curl -s http://169.254.169.254/latest/meta-data/local-ipv4)
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4 || echo "none")

echo "Instance ID: $INSTANCE_ID"
echo "Availability Zone: $AVAILABILITY_ZONE"
echo "Private IP: $PRIVATE_IP"
echo "Public IP: $PUBLIC_IP"

# Install K3s agent
echo "Installing K3s agent..."
export K3S_URL="%s"
export K3S_TOKEN="%s"
export K3S_NODE_NAME="%s"
export INSTALL_K3S_EXEC="agent --node-label %s --node-taint %s"

# Download and install K3s
curl -sfL https://get.k3s.io | sh -

# Wait for K3s to be ready
echo "Waiting for K3s agent to start..."
for i in {1..30}; do
  if systemctl is-active --quiet k3s-agent; then
    echo "K3s agent is running"
    break
  fi
  echo "Waiting for K3s agent to start... ($i/30)"
  sleep 5
done

# Check K3s agent status
if systemctl is-active --quiet k3s-agent; then
  echo "K3s agent successfully started"
  systemctl status k3s-agent --no-pager
else
  echo "ERROR: K3s agent failed to start"
  systemctl status k3s-agent --no-pager || true
  journalctl -u k3s-agent -n 50 --no-pager || true
  exit 1
fi

# Enable K3s agent to start on boot
echo "Enabling K3s agent to start on boot..."
systemctl enable k3s-agent

# Install additional tools
echo "Installing additional tools..."
apt-get install -y git wget htop net-tools || yum install -y git wget htop net-tools || true

# Setup SSH access for workspaces
echo "Configuring SSH..."
mkdir -p /root/.ssh
chmod 700 /root/.ssh

# Add SSH public keys from WorkMachine spec
%s

# Optimize network settings for workspaces
echo "Optimizing network settings..."
cat >> /etc/sysctl.conf <<EOF
# Increase network buffers for better performance
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.tcp_rmem = 4096 87380 67108864
net.ipv4.tcp_wmem = 4096 65536 67108864

# Increase connection tracking
net.netfilter.nf_conntrack_max = 1048576
net.nf_conntrack_max = 1048576

# Enable BBR congestion control
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr
EOF
sysctl -p || true

# Setup workspace home directories
echo "Creating workspace directories..."
mkdir -p /var/lib/workspaces
chmod 755 /var/lib/workspaces

# Create Nix store directory for shared packages
echo "Creating Nix store directory..."
mkdir -p /nix/store
chmod 755 /nix

# Install Docker (required for some workspaces)
echo "Installing Docker..."
curl -fsSL https://get.docker.com | sh || true
systemctl enable docker || true
systemctl start docker || true

echo "K3s agent installation completed successfully"
echo "WorkMachine %s is ready"
echo "Timestamp: $(date)"

# Signal completion by creating a marker file
touch /var/lib/k3s-install-complete
`, wm.Name, awsConfig.K3sServerURL, k3sToken, wm.Name,
		strings.Join(nodeLabels, ","),
		strings.Join(nodeTaints, ","),
		generateSSHKeysSection(wm.Spec.SSHPublicKeys),
		wm.Name)

	// Base64 encode the user data (AWS expects base64 for user data)
	return base64.StdEncoding.EncodeToString([]byte(userDataScript)), nil
}

// generateSSHKeysSection generates the SSH keys section for user data
func generateSSHKeysSection(sshPublicKeys []string) string {
	if len(sshPublicKeys) == 0 {
		return "# No SSH keys configured"
	}

	var sb strings.Builder
	sb.WriteString("# Add user SSH public keys\n")
	for _, key := range sshPublicKeys {
		if key != "" {
			sb.WriteString(fmt.Sprintf("echo '%s' >> /root/.ssh/authorized_keys\n", key))
		}
	}
	sb.WriteString("chmod 600 /root/.ssh/authorized_keys\n")
	return sb.String()
}
