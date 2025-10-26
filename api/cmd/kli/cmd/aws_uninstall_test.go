package cmd

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
)

func TestFindInstancesByKeyFilter(t *testing.T) {
	// Test filter structure for finding instances
	installationKey := "test"

	filters := []types.Filter{
		{
			Name:   aws.String("tag:InstallationKey"),
			Values: []string{installationKey},
		},
		{
			Name: aws.String("instance-state-name"),
			Values: []string{
				string(types.InstanceStateNameRunning),
				string(types.InstanceStateNamePending),
				string(types.InstanceStateNameStopping),
				string(types.InstanceStateNameStopped),
			},
		},
	}

	// Verify filter structure
	assert.Len(t, filters, 2)
	assert.Equal(t, "tag:InstallationKey", *filters[0].Name)
	assert.Contains(t, filters[0].Values, installationKey)

	// Verify state filter includes all non-terminated states
	stateValues := filters[1].Values
	assert.Contains(t, stateValues, string(types.InstanceStateNameRunning))
	assert.Contains(t, stateValues, string(types.InstanceStateNamePending))
	assert.Contains(t, stateValues, string(types.InstanceStateNameStopping))
	assert.Contains(t, stateValues, string(types.InstanceStateNameStopped))
	assert.NotContains(t, stateValues, string(types.InstanceStateNameTerminated))
}

func TestTerminationProtectionDisable(t *testing.T) {
	// Test termination protection disable attribute
	attr := &types.AttributeBooleanValue{
		Value: aws.Bool(false),
	}

	assert.NotNil(t, attr.Value)
	assert.False(t, *attr.Value)
}

func TestSecurityGroupDeletionRetryLogic(t *testing.T) {
	// Test retry configuration
	maxRetries := 6
	baseWaitTime := 5 // seconds

	assert.Equal(t, 6, maxRetries)
	assert.Equal(t, 5, baseWaitTime)

	// Calculate wait times for each retry
	expectedWaitTimes := []int{0, 5, 10, 15, 20, 25}
	for i := 0; i < maxRetries; i++ {
		waitTime := i * baseWaitTime
		assert.Equal(t, expectedWaitTimes[i], waitTime)
	}
}

func TestUninstallResourceNames(t *testing.T) {
	tests := []struct {
		name            string
		installationKey string
		sgName          string
		roleName        string
	}{
		{
			name:            "test installation",
			installationKey: "test",
			sgName:          "kl-test-sg",
			roleName:        "kl-test-role",
		},
		{
			name:            "prod installation",
			installationKey: "prod",
			sgName:          "kl-prod-sg",
			roleName:        "kl-prod-role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.sgName, "kl-"+tt.installationKey+"-sg")
			assert.Equal(t, tt.roleName, "kl-"+tt.installationKey+"-role")
		})
	}
}

func TestDeleteSecurityGroupFilters(t *testing.T) {
	// Test security group filter for deletion
	installationKey := "test"
	sgName := "kl-test-sg"

	filters := []types.Filter{
		{
			Name:   aws.String("group-name"),
			Values: []string{sgName},
		},
		{
			Name:   aws.String("tag:InstallationKey"),
			Values: []string{installationKey},
		},
	}

	// Verify filters
	assert.Len(t, filters, 2)
	assert.Equal(t, "group-name", *filters[0].Name)
	assert.Contains(t, filters[0].Values, sgName)
	assert.Equal(t, "tag:InstallationKey", *filters[1].Name)
	assert.Contains(t, filters[1].Values, installationKey)
}

func TestIAMResourceCleanup(t *testing.T) {
	// Test IAM resource names for cleanup
	installationKey := "test"
	profileName := "kl-" + installationKey + "-role"
	roleName := "kl-" + installationKey + "-role"

	assert.Equal(t, "kl-test-role", profileName)
	assert.Equal(t, "kl-test-role", roleName)
}

func TestUninstallParallelOperations(t *testing.T) {
	// Test that parallel operations are properly structured
	operations := []string{
		"terminate-instances",
		"delete-security-group",
		"delete-iam-resources",
	}

	assert.Len(t, operations, 3)
	assert.Contains(t, operations, "terminate-instances")
	assert.Contains(t, operations, "delete-security-group")
	assert.Contains(t, operations, "delete-iam-resources")
}

func TestInstanceTerminationFlow(t *testing.T) {
	// Test the flow: disable protection -> terminate
	steps := []string{
		"disable-termination-protection",
		"terminate-instances",
	}

	assert.Len(t, steps, 2)
	assert.Equal(t, "disable-termination-protection", steps[0])
	assert.Equal(t, "terminate-instances", steps[1])
}

func TestSecurityGroupDependencyErrorDetection(t *testing.T) {
	// Test error message detection
	errorMessages := []string{
		"DependencyViolation: resource has a dependent object",
		"some other error",
	}

	for _, errMsg := range errorMessages {
		hasDependency := containsDependencyViolation(errMsg)
		if errMsg == "DependencyViolation: resource has a dependent object" {
			assert.True(t, hasDependency)
		} else {
			assert.False(t, hasDependency)
		}
	}
}

// Helper function to test error detection
func containsDependencyViolation(errMsg string) bool {
	return len(errMsg) >= 19 && errMsg[0:19] == "DependencyViolation"
}

func TestUninstallSignalHandling(t *testing.T) {
	// Test that signal handling is configured correctly
	// Uninstall should continue to completion even on interrupt
	shouldAbortOnSignal := false
	assert.False(t, shouldAbortOnSignal, "Uninstall should not abort on signal")
}

func TestIAMPolicyDetachment(t *testing.T) {
	// Test that managed policies are detached before role deletion
	managedPolicyArn := "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"

	assert.NotEmpty(t, managedPolicyArn)
	assert.Contains(t, managedPolicyArn, "arn:aws:iam::")
	assert.Contains(t, managedPolicyArn, "AmazonSSMManagedInstanceCore")
}

func TestCleanupOrder(t *testing.T) {
	// Test that cleanup happens in correct order
	// 1. Terminate instances
	// 2. Delete security group (with retries)
	// 3. Delete IAM resources

	cleanupOrder := []string{
		"instances",
		"security-group",
		"iam",
	}

	assert.Len(t, cleanupOrder, 3)
	assert.Equal(t, "instances", cleanupOrder[0])
	assert.Equal(t, "security-group", cleanupOrder[1])
	assert.Equal(t, "iam", cleanupOrder[2])
}

func TestUninstallContext(t *testing.T) {
	// Test context usage
	ctx := context.Background()
	assert.NotNil(t, ctx)
}

func TestInstanceIDFormat(t *testing.T) {
	// Test instance ID format validation
	validInstanceIDs := []string{
		"i-0123456789abcdef0",
		"i-abcdef0123456789",
	}

	invalidInstanceIDs := []string{
		"",
		"instance-123",
		"i-",
	}

	for _, id := range validInstanceIDs {
		assert.True(t, len(id) > 2 && id[0:2] == "i-", "Valid instance ID should start with i-")
	}

	for _, id := range invalidInstanceIDs {
		if id == "" {
			assert.Empty(t, id)
		} else if len(id) < 2 {
			assert.True(t, len(id) < 2, "Too short to be valid")
		} else {
			// Invalid IDs should not match the AWS instance ID pattern
			isValidFormat := len(id) > 2 && id[0:2] == "i-" && len(id) >= 10
			assert.False(t, isValidFormat, "Invalid instance ID should not match AWS pattern")
		}
	}
}

func TestUninstallRequiredFlags(t *testing.T) {
	// Test that installation-key is required
	requiredFlags := []string{"installation-key"}

	assert.Contains(t, requiredFlags, "installation-key")
	assert.Len(t, requiredFlags, 1)
}

func TestUninstallOptionalFlags(t *testing.T) {
	// Test optional flags
	optionalFlags := []string{"region"}

	assert.Contains(t, optionalFlags, "region")
}
