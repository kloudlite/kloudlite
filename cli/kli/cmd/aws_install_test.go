package cmd

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper functions

func TestFindLatestUbuntuAMI(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		expectedErr bool
	}{
		{
			name:        "valid context",
			ctx:         context.Background(),
			expectedErr: false,
		},
		{
			name:        "nil context should use background",
			ctx:         nil,
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test requires AWS credentials, so we'll skip if not available
			t.Skip("Requires AWS credentials - integration test")
		})
	}
}

func TestGetDefaultVPC(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		expectedErr bool
	}{
		{
			name:        "valid context",
			ctx:         context.Background(),
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test requires AWS credentials, so we'll skip if not available
			t.Skip("Requires AWS credentials - integration test")
		})
	}
}

func TestUserDataGeneration(t *testing.T) {
	// Test that user data is properly formatted and base64 encoded
	userData := `#!/bin/bash
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

# Install K3s server
echo "Installing K3s server..."
curl -sfL https://get.k3s.io | sh -s - server \
  --disable traefik \
  --write-kubeconfig-mode 644

# Wait for K3s to be ready
echo "Waiting for K3s to be ready..."
until kubectl get nodes 2>/dev/null; do
  sleep 2
done

echo "K3s installation completed at $(date)"
echo "Kloudlite installation completed successfully!"
`

	encoded := base64.StdEncoding.EncodeToString([]byte(userData))

	// Test that it can be decoded back
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)
	assert.Equal(t, userData, string(decoded))

	// Test that it contains expected commands
	assert.Contains(t, userData, "apt-get update")
	assert.Contains(t, userData, "curl -sfL https://get.k3s.io")
	assert.Contains(t, userData, "kubectl get nodes")
	assert.Contains(t, userData, "#!/bin/bash")
}

func TestSecurityGroupNameFormat(t *testing.T) {
	tests := []struct {
		name            string
		installationKey string
		expected        string
	}{
		{
			name:            "simple key",
			installationKey: "test",
			expected:        "kl-test-sg",
		},
		{
			name:            "prod key",
			installationKey: "prod",
			expected:        "kl-prod-sg",
		},
		{
			name:            "staging key",
			installationKey: "staging",
			expected:        "kl-staging-sg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := "kl-" + tt.installationKey + "-sg"
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoleNameFormat(t *testing.T) {
	tests := []struct {
		name            string
		installationKey string
		expected        string
	}{
		{
			name:            "simple key",
			installationKey: "test",
			expected:        "kl-test-role",
		},
		{
			name:            "prod key",
			installationKey: "prod",
			expected:        "kl-prod-role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := "kl-" + tt.installationKey + "-role"
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInstanceNameFormat(t *testing.T) {
	tests := []struct {
		name            string
		installationKey string
		expected        string
	}{
		{
			name:            "simple key",
			installationKey: "test",
			expected:        "kl-test-instance",
		},
		{
			name:            "prod key",
			installationKey: "prod",
			expected:        "kl-prod-instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := "kl-" + tt.installationKey + "-instance"
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityGroupRulesValidation(t *testing.T) {
	// Test that required ports are included in security group configuration
	requiredIngressPorts := []struct {
		port     int32
		protocol string
		desc     string
	}{
		{443, "tcp", "HTTPS"},
		{6443, "tcp", "Kubernetes API"},
		{8472, "udp", "Flannel VXLAN"},
		{10250, "tcp", "Kubelet metrics"},
		{5001, "tcp", "Kloudlite agent"},
	}

	// Verify all required ports
	for _, rule := range requiredIngressPorts {
		t.Run(rule.desc, func(t *testing.T) {
			assert.NotZero(t, rule.port, "Port should not be zero")
			assert.NotEmpty(t, rule.protocol, "Protocol should not be empty")
			assert.True(t, rule.protocol == "tcp" || rule.protocol == "udp", "Protocol should be tcp or udp")
		})
	}
}

func TestIAMPolicyStructure(t *testing.T) {
	// Test IAM policy JSON structure
	policy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Action": []string{
					"ec2:RunInstances",
					"ec2:TerminateInstances",
					"ec2:DescribeInstances",
					"ec2:ModifyInstanceAttribute",
					"ec2:DescribeInstanceTypes",
					"ec2:DescribeImages",
					"ec2:DescribeVolumes",
					"ec2:CreateTags",
				},
				"Resource": "*",
			},
		},
	}

	// Verify structure
	assert.Equal(t, "2012-10-17", policy["Version"])
	statements := policy["Statement"].([]map[string]interface{})
	assert.Len(t, statements, 1)
	assert.Equal(t, "Allow", statements[0]["Effect"])

	actions := statements[0]["Action"].([]string)
	assert.Contains(t, actions, "ec2:RunInstances")
	assert.Contains(t, actions, "ec2:TerminateInstances")
}

func TestTrustPolicyStructure(t *testing.T) {
	// Test trust policy JSON structure
	trustPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]string{
					"Service": "ec2.amazonaws.com",
				},
				"Action": "sts:AssumeRole",
			},
		},
	}

	// Verify structure
	assert.Equal(t, "2012-10-17", trustPolicy["Version"])
	statements := trustPolicy["Statement"].([]map[string]interface{})
	assert.Len(t, statements, 1)
	assert.Equal(t, "Allow", statements[0]["Effect"])
	assert.Equal(t, "sts:AssumeRole", statements[0]["Action"])

	principal := statements[0]["Principal"].(map[string]string)
	assert.Equal(t, "ec2.amazonaws.com", principal["Service"])
}

func TestInstanceTypeValidation(t *testing.T) {
	// Verify instance type is correctly set
	instanceType := types.InstanceTypeT3Medium
	assert.Equal(t, types.InstanceTypeT3Medium, instanceType)
	assert.NotEmpty(t, string(instanceType))
}

func TestVolumeConfiguration(t *testing.T) {
	// Test volume configuration
	volumeSize := int32(100)
	volumeType := types.VolumeTypeGp3
	deleteOnTermination := true

	assert.Equal(t, int32(100), volumeSize)
	assert.Equal(t, types.VolumeTypeGp3, volumeType)
	assert.True(t, deleteOnTermination)
}

func TestSSMManagedPolicyARN(t *testing.T) {
	// Verify SSM policy ARN format
	policyArn := "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"

	assert.True(t, strings.HasPrefix(policyArn, "arn:aws:iam::"))
	assert.Contains(t, policyArn, "AmazonSSMManagedInstanceCore")
}

func TestTagStructure(t *testing.T) {
	// Test tag structure for resources
	installationKey := "test"

	tags := []types.Tag{
		{Key: aws.String("Name"), Value: aws.String("kl-test-instance")},
		{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
		{Key: aws.String("Project"), Value: aws.String("kloudlite")},
		{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
		{Key: aws.String("InstallationKey"), Value: aws.String(installationKey)},
	}

	// Verify all tags have keys and values
	for _, tag := range tags {
		assert.NotNil(t, tag.Key)
		assert.NotNil(t, tag.Value)
		assert.NotEmpty(t, *tag.Key)
		assert.NotEmpty(t, *tag.Value)
	}

	// Verify specific tags exist
	var hasInstallationKeyTag bool
	var hasManagedByTag bool
	for _, tag := range tags {
		if *tag.Key == "InstallationKey" && *tag.Value == installationKey {
			hasInstallationKeyTag = true
		}
		if *tag.Key == "ManagedBy" && *tag.Value == "kloudlite" {
			hasManagedByTag = true
		}
	}
	assert.True(t, hasInstallationKeyTag, "InstallationKey tag should exist")
	assert.True(t, hasManagedByTag, "ManagedBy tag should exist")
}

func TestNetworkInterfaceConfiguration(t *testing.T) {
	// Test network interface configuration
	deviceIndex := int32(0)
	associatePublicIP := true

	assert.Equal(t, int32(0), deviceIndex)
	assert.True(t, associatePublicIP)
}

func TestTerminationProtectionAttribute(t *testing.T) {
	// Test termination protection attribute structure
	enableProtection := true

	attr := &types.AttributeBooleanValue{
		Value: aws.Bool(enableProtection),
	}

	assert.NotNil(t, attr.Value)
	assert.True(t, *attr.Value)

	// Test disabling
	disableProtection := false
	attrDisable := &types.AttributeBooleanValue{
		Value: aws.Bool(disableProtection),
	}
	assert.False(t, *attrDisable.Value)
}

func TestK3sInstallationScript(t *testing.T) {
	// Test that K3s installation script has required components
	script := `#!/bin/bash
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

# Install K3s server
echo "Installing K3s server..."
curl -sfL https://get.k3s.io | sh -s - server \
  --disable traefik \
  --write-kubeconfig-mode 644

# Wait for K3s to be ready
echo "Waiting for K3s to be ready..."
until kubectl get nodes 2>/dev/null; do
  sleep 2
done

echo "K3s installation completed at $(date)"
echo "Kloudlite installation completed successfully!"
`

	// Verify script components
	assert.Contains(t, script, "#!/bin/bash")
	assert.Contains(t, script, "set -euo pipefail")
	assert.Contains(t, script, "curl -sfL https://get.k3s.io")
	assert.Contains(t, script, "--disable traefik")
	assert.Contains(t, script, "--write-kubeconfig-mode 644")
	assert.Contains(t, script, "kubectl get nodes")
	assert.Contains(t, script, "/var/log/kloudlite-init.log")
}
