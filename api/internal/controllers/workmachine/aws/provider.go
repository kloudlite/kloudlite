package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	route53types "github.com/aws/aws-sdk-go-v2/service/route53/types"

	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/errors"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/types"
	workmachinev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
)

// Provider implements the CloudProviderInterface for AWS
type Provider struct {
	client   *Client
	config   *workmachinev1.AWSProviderConfig
	k3sToken string
}

// NewProvider creates a new AWS provider with initialized client
func NewProvider(ctx context.Context, cfg *workmachinev1.AWSProviderConfig, k3sToken string) (*Provider, error) {
	if cfg == nil {
		return nil, errors.NewInvalidConfigurationError("aws", "AWS configuration is required")
	}

	if cfg.Region == "" {
		return nil, errors.NewInvalidConfigurationError("region", "region is required")
	}

	// Create AWS client immediately
	client, err := NewClient(ctx, cfg.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %w", err)
	}

	return &Provider{
		client:   client,
		config:   cfg,
		k3sToken: k3sToken,
	}, nil
}

// ValidatePermissions validates that the AWS credentials have all required permissions
func (p *Provider) ValidatePermissions(ctx context.Context) error {
	return ValidatePermissions(ctx, p.client, p.config.VPCID, p.config.SubnetID, p.config.Route53HostedZoneID)
}

// CreateInstance creates a new EC2 instance for the WorkMachine
// Returns ResourceAlreadyExistsError if an instance with the same tags already exists
func (p *Provider) CreateInstance(ctx context.Context, wm *workmachinev1.WorkMachine) (*types.InstanceInfo, error) {
	// Generate user data using stored K3s token
	userData, err := GenerateK3sUserData(wm, p.k3sToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate user data: %w", err)
	}

	// Get security group ID (must be created before calling this)
	sgID := wm.Status.SecurityGroupID
	if sgID == "" {
		return nil, errors.NewInvalidConfigurationError("securityGroupID", "security group must be created before instance")
	}

	// Build security group IDs list
	securityGroupIDs := []string{sgID}
	if len(p.config.SecurityGroupIDs) > 0 {
		securityGroupIDs = append(securityGroupIDs, p.config.SecurityGroupIDs...)
	}

	// Prepare instance tags
	tags := []ec2types.Tag{
		{Key: fn.Ptr("Name"), Value: fn.Ptr(fmt.Sprintf("workmachine-%s", wm.Name))},
		{Key: fn.Ptr("kloudlite.io/workmachine"), Value: fn.Ptr(wm.Name)},
		{Key: fn.Ptr("kloudlite.io/owner"), Value: fn.Ptr(wm.Spec.OwnedBy)},
		{Key: fn.Ptr("kloudlite.io/managed-by"), Value: fn.Ptr("kloudlite-controller")},
	}

	// Prepare IAM instance profile
	var iamInstanceProfile *ec2types.IamInstanceProfileSpecification
	if p.config.IAMInstanceProfileRole != nil && *p.config.IAMInstanceProfileRole != "" {
		iamInstanceProfile = &ec2types.IamInstanceProfileSpecification{
			Name: p.config.IAMInstanceProfileRole,
		}
	}

	// Create EC2 instance
	runInput := &ec2.RunInstancesInput{
		ImageId:            fn.Ptr(p.config.AMI),
		InstanceType:       ec2types.InstanceType(p.config.InstanceType),
		MinCount:           fn.Ptr(int32(1)),
		MaxCount:           fn.Ptr(int32(1)),
		SubnetId:           fn.Ptr(p.config.SubnetID),
		SecurityGroupIds:   securityGroupIDs,
		UserData:           fn.Ptr(userData),
		IamInstanceProfile: iamInstanceProfile,
		BlockDeviceMappings: []ec2types.BlockDeviceMapping{
			{
				DeviceName: fn.Ptr("/dev/sda1"),
				Ebs: &ec2types.EbsBlockDevice{
					VolumeSize:          fn.Ptr(int32(p.config.RootVolumeSize)),
					VolumeType:          ec2types.VolumeType(p.config.RootVolumeType),
					DeleteOnTermination: fn.Ptr(true),
				},
			},
		},
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeInstance,
				Tags:         tags,
			},
			{
				ResourceType: ec2types.ResourceTypeVolume,
				Tags:         tags,
			},
		},
	}

	// Add availability zone if specified
	if p.config.AvailabilityZone != "" {
		runInput.Placement = &ec2types.Placement{
			AvailabilityZone: fn.Ptr(p.config.AvailabilityZone),
		}
	}

	runOutput, err := p.client.EC2.RunInstances(ctx, runInput)
	if err != nil {
		return nil, errors.NewProviderAPIError("RunInstances", "", fmt.Sprintf("failed to create instance: %v", err), true, err)
	}

	if len(runOutput.Instances) == 0 {
		return nil, fmt.Errorf("no instances were created")
	}

	instance := runOutput.Instances[0]

	return &types.InstanceInfo{
		InstanceID:       aws.ToString(instance.InstanceId),
		State:            mapEC2StateToInstanceState(instance.State),
		PrivateIP:        aws.ToString(instance.PrivateIpAddress),
		PublicIP:         aws.ToString(instance.PublicIpAddress),
		Region:           p.config.Region,
		AvailabilityZone: aws.ToString(instance.Placement.AvailabilityZone),
		SecurityGroupID:  sgID,
		Message:          "Instance created successfully",
	}, nil
}

// GetInstance retrieves the current status of an EC2 instance by its ID
func (p *Provider) GetInstance(ctx context.Context, instanceID string) (*types.InstanceInfo, error) {
	if instanceID == "" {
		return nil, errors.NewResourceNotFoundError("instance", "")
	}

	output, err := p.client.EC2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		if strings.Contains(err.Error(), "InvalidInstanceID.NotFound") {
			return nil, errors.NewResourceNotFoundError("instance", instanceID)
		}
		return nil, errors.NewProviderAPIError("DescribeInstances", "", fmt.Sprintf("failed to describe instance: %v", err), true, err)
	}

	if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return nil, errors.NewResourceNotFoundError("instance", instanceID)
	}

	instance := output.Reservations[0].Instances[0]

	// Get security group ID
	var sgID string
	if len(instance.SecurityGroups) > 0 {
		sgID = aws.ToString(instance.SecurityGroups[0].GroupId)
	}

	return &types.InstanceInfo{
		InstanceID:       aws.ToString(instance.InstanceId),
		State:            mapEC2StateToInstanceState(instance.State),
		PrivateIP:        aws.ToString(instance.PrivateIpAddress),
		PublicIP:         aws.ToString(instance.PublicIpAddress),
		Region:           p.config.Region,
		AvailabilityZone: aws.ToString(instance.Placement.AvailabilityZone),
		SecurityGroupID:  sgID,
		Message:          fmt.Sprintf("Instance is %s", string(instance.State.Name)),
	}, nil
}

// StartInstance starts a stopped EC2 instance
func (p *Provider) StartInstance(ctx context.Context, instanceID string) error {
	if instanceID == "" {
		return errors.NewResourceNotFoundError("instance", "")
	}

	_, err := p.client.EC2.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		if strings.Contains(err.Error(), "InvalidInstanceID.NotFound") {
			return errors.NewResourceNotFoundError("instance", instanceID)
		}
		return errors.NewProviderAPIError("StartInstances", "", fmt.Sprintf("failed to start instance: %v", err), true, err)
	}

	return nil
}

// StopInstance stops a running EC2 instance
func (p *Provider) StopInstance(ctx context.Context, instanceID string) error {
	if instanceID == "" {
		return errors.NewResourceNotFoundError("instance", "")
	}

	_, err := p.client.EC2.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		if strings.Contains(err.Error(), "InvalidInstanceID.NotFound") {
			return errors.NewResourceNotFoundError("instance", instanceID)
		}
		return errors.NewProviderAPIError("StopInstances", "", fmt.Sprintf("failed to stop instance: %v", err), true, err)
	}

	return nil
}

// DeleteInstance permanently deletes an EC2 instance
func (p *Provider) DeleteInstance(ctx context.Context, instanceID string) error {
	if instanceID == "" {
		return nil // Nothing to delete
	}

	// Terminate the instance
	_, err := p.client.EC2.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{instanceID},
	})
	if err != nil {
		// Check if instance is already terminated or doesn't exist
		if strings.Contains(err.Error(), "InvalidInstanceID.NotFound") {
			return nil
		}
		return errors.NewProviderAPIError("TerminateInstances", "", fmt.Sprintf("failed to terminate instance: %v", err), true, err)
	}

	return nil
}

// UpsertDNSRecord creates or updates a DNS record in Route53
// fqdn is the fully qualified domain name (controller constructs this)
func (p *Provider) UpsertDNSRecord(ctx context.Context, fqdn, publicIP string) error {
	if publicIP == "" {
		return errors.NewInvalidConfigurationError("publicIP", "public IP is required for DNS configuration")
	}

	if fqdn == "" {
		return errors.NewInvalidConfigurationError("fqdn", "FQDN is required for DNS configuration")
	}

	// Ensure trailing dot for Route53
	if !strings.HasSuffix(fqdn, ".") {
		fqdn = fqdn + "."
	}

	// Create or update the DNS record
	changeInput := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: fn.Ptr(p.config.Route53HostedZoneID),
		ChangeBatch: &route53types.ChangeBatch{
			Changes: []route53types.Change{
				{
					Action: route53types.ChangeActionUpsert,
					ResourceRecordSet: &route53types.ResourceRecordSet{
						Name: fn.Ptr(fqdn),
						Type: route53types.RRTypeA,
						TTL:  fn.Ptr(int64(300)),
						ResourceRecords: []route53types.ResourceRecord{
							{Value: fn.Ptr(publicIP)},
						},
					},
				},
			},
			Comment: fn.Ptr(fmt.Sprintf("WorkMachine DNS record for %s", fqdn)),
		},
	}

	_, err := p.client.Route53.ChangeResourceRecordSets(ctx, changeInput)
	if err != nil {
		return errors.NewDNSError(fqdn, "upsert", err)
	}

	return nil
}

// DeleteDNSRecord removes a DNS record from Route53
// fqdn is the fully qualified domain name (controller provides this)
func (p *Provider) DeleteDNSRecord(ctx context.Context, fqdn, publicIP string) error {
	if fqdn == "" {
		return nil // Nothing to delete
	}

	if publicIP == "" {
		return nil // Can't delete without knowing the IP
	}

	// Ensure trailing dot for Route53
	if !strings.HasSuffix(fqdn, ".") {
		fqdn = fqdn + "."
	}

	// Delete the DNS record
	changeInput := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: fn.Ptr(p.config.Route53HostedZoneID),
		ChangeBatch: &route53types.ChangeBatch{
			Changes: []route53types.Change{
				{
					Action: route53types.ChangeActionDelete,
					ResourceRecordSet: &route53types.ResourceRecordSet{
						Name: fn.Ptr(fqdn),
						Type: route53types.RRTypeA,
						TTL:  fn.Ptr(int64(300)),
						ResourceRecords: []route53types.ResourceRecord{
							{Value: fn.Ptr(publicIP)},
						},
					},
				},
			},
		},
	}

	_, err := p.client.Route53.ChangeResourceRecordSets(ctx, changeInput)
	if err != nil {
		// If record doesn't exist, that's fine
		if strings.Contains(err.Error(), "it was not found") {
			return nil
		}
		return errors.NewDNSError(fqdn, "delete", err)
	}

	return nil
}

// CreateSecurityGroup creates a new security group for the WorkMachine
// Returns ResourceAlreadyExistsError if a security group with the same name already exists
func (p *Provider) CreateSecurityGroup(ctx context.Context, wm *workmachinev1.WorkMachine) (string, error) {
	sgName := fmt.Sprintf("workmachine-%s", wm.Name)
	sgDescription := fmt.Sprintf("Security group for WorkMachine %s (owner: %s)", wm.Name, wm.Spec.OwnedBy)

	// Create security group
	createOutput, err := p.client.EC2.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   fn.Ptr(sgName),
		Description: fn.Ptr(sgDescription),
		VpcId:       fn.Ptr(p.config.VPCID),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeSecurityGroup,
				Tags: []ec2types.Tag{
					{Key: fn.Ptr("Name"), Value: fn.Ptr(sgName)},
					{Key: fn.Ptr("kloudlite.io/workmachine"), Value: fn.Ptr(wm.Name)},
					{Key: fn.Ptr("kloudlite.io/owner"), Value: fn.Ptr(wm.Spec.OwnedBy)},
				},
			},
		},
	})
	if err != nil {
		// Check if security group already exists
		if strings.Contains(err.Error(), "InvalidGroup.Duplicate") {
			return "", errors.NewResourceAlreadyExistsError("security-group", sgName)
		}
		return "", fmt.Errorf("failed to create security group: %w", err)
	}

	sgID := aws.ToString(createOutput.GroupId)

	// Add ingress rule for port 443 (HTTPS) and port 22 (SSH)
	_, err = p.client.EC2.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: fn.Ptr(sgID),
		IpPermissions: []ec2types.IpPermission{
			{
				IpProtocol: fn.Ptr("tcp"),
				FromPort:   fn.Ptr(int32(443)),
				ToPort:     fn.Ptr(int32(443)),
				IpRanges: []ec2types.IpRange{
					{
						CidrIp:      fn.Ptr("0.0.0.0/0"),
						Description: fn.Ptr("HTTPS access to workspace services"),
					},
				},
			},
			{
				IpProtocol: fn.Ptr("tcp"),
				FromPort:   fn.Ptr(int32(22)),
				ToPort:     fn.Ptr(int32(22)),
				IpRanges: []ec2types.IpRange{
					{
						CidrIp:      fn.Ptr("0.0.0.0/0"),
						Description: fn.Ptr("SSH access for debugging"),
					},
				},
			},
		},
	})
	if err != nil {
		// If the rule already exists, that's fine
		if !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
			return "", fmt.Errorf("failed to authorize security group ingress: %w", err)
		}
	}

	// Egress rules are automatically created by AWS (allow all outbound traffic)
	// But we'll ensure it explicitly
	_, err = p.client.EC2.AuthorizeSecurityGroupEgress(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId: fn.Ptr(sgID),
		IpPermissions: []ec2types.IpPermission{
			{
				IpProtocol: fn.Ptr("-1"), // All protocols
				IpRanges: []ec2types.IpRange{
					{
						CidrIp:      fn.Ptr("0.0.0.0/0"),
						Description: fn.Ptr("Allow all outbound traffic"),
					},
				},
			},
		},
	})
	if err != nil {
		// If the rule already exists, that's fine
		if !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
			return "", fmt.Errorf("failed to authorize security group egress: %w", err)
		}
	}

	return sgID, nil
}

// GetSecurityGroup retrieves a security group by name
// Returns empty string if not found (not an error)
func (p *Provider) GetSecurityGroup(ctx context.Context, name string) (string, error) {
	describeOutput, err := p.client.EC2.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: []ec2types.Filter{
			{
				Name:   fn.Ptr("group-name"),
				Values: []string{name},
			},
			{
				Name:   fn.Ptr("vpc-id"),
				Values: []string{p.config.VPCID},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe security groups: %w", err)
	}

	// If security group exists, return its ID
	if len(describeOutput.SecurityGroups) > 0 {
		return aws.ToString(describeOutput.SecurityGroups[0].GroupId), nil
	}

	return "", nil // Not found, but not an error
}

// DeleteSecurityGroup deletes a security group by ID
func (p *Provider) DeleteSecurityGroup(ctx context.Context, sgID string) error {
	if sgID == "" {
		return nil // Nothing to delete
	}

	_, err := p.client.EC2.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: fn.Ptr(sgID),
	})
	if err != nil {
		if strings.Contains(err.Error(), "InvalidGroup.NotFound") {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete security group: %w", err)
	}
	return nil
}

// mapEC2StateToInstanceState maps EC2 instance state to InstanceState
func mapEC2StateToInstanceState(state *ec2types.InstanceState) types.InstanceState {
	if state == nil {
		return types.InstanceStateNotFound
	}

	switch state.Name {
	case ec2types.InstanceStateNamePending:
		return types.InstanceStatePending
	case ec2types.InstanceStateNameRunning:
		return types.InstanceStateRunning
	case ec2types.InstanceStateNameStopping:
		return types.InstanceStateStopping
	case ec2types.InstanceStateNameStopped:
		return types.InstanceStateStopped
	case ec2types.InstanceStateNameShuttingDown:
		return types.InstanceStateTerminating
	case ec2types.InstanceStateNameTerminated:
		return types.InstanceStateTerminated
	default:
		return types.InstanceStateError
	}
}
