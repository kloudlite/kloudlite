package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/templates"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"golang.org/x/sync/errgroup"
)

type provider struct {
	ec2Client *ec2.Client
	k3sURL    string
	k3sToken  string

	// ---
	region string
	vpcID  *string
}

var _ cloud.Provider = (*provider)(nil)

func NewProvider(ctx context.Context, k3sURL string, k3sToken string) (cloud.Provider, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithEC2IMDSRegion())
	if err != nil {
		return nil, errors.Wrap("failed to load AWS config", err)
	}

	slog.Info("aws default config", "region", awsCfg.Region)

	ec2Client := ec2.NewFromConfig(awsCfg)

	defaultVPC, err := ec2Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{
		Filters: []ec2types.Filter{
			{
				Name:   fn.Ptr("isDefault"),
				Values: []string{"true"},
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap("failed to get current client VPC", err)
	}

	slog.Info("post describe VPCs call", "defaultVPC", defaultVPC.Vpcs)

	if len(defaultVPC.Vpcs) != 1 {
		return nil, errors.New("got no results for default VPC")
	}

	return &provider{
		ec2Client: ec2.NewFromConfig(awsCfg),

		k3sURL:   k3sURL,
		k3sToken: k3sToken,

		region: awsCfg.Region,
		vpcID:  defaultVPC.Vpcs[0].VpcId,
	}, nil
}

func handleDryRunError(err error, action string) error {
	if err == nil {
		return fmt.Errorf("dry-run for %s should have failed with DryRunOperation", action)
	}

	// Check if it's a DryRunOperation error (indicates permission is granted)
	if strings.Contains(err.Error(), "DryRunOperation") || strings.Contains(err.Error(), "Request would have succeeded") {
		return nil
	}

	return err
}

func (p *provider) ValidatePermissions(ctx context.Context) error {
	g, errctx := errgroup.WithContext(ctx)
	dryRunChecks := []func(context.Context) error{
		p.dryRunCreateInstance,
	}

	for i := range dryRunChecks {
		g.Go(func() error {
			return dryRunChecks[i](errctx)
		})
	}

	return g.Wait()
}

func (p *provider) CreateMachine(ctx context.Context, wm *v1.WorkMachine) (*v1.MachineInfo, error) {
	// Step 2: Get or create security group
	sgID, err := p.getOrCreateSecurityGroup(ctx, wm.Spec.AWSProvider.VPC_ID, wm.Spec.AllowedCIDR)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create security group: %w", err)
	}

	sshKeyPair, err := p.getOrCreateSSHKeyPair(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create SSH key pair: %w", err)
	}

	// Step 3: Build security group IDs list
	securityGroupIDs := []string{sgID}

	// Step 4: Prepare instance tags
	tags := []ec2types.Tag{
		{Key: fn.Ptr("Name"), Value: fn.Ptr("workmachine-" + wm.Name)},
		{Key: fn.Ptr("kloudlite.io/workmachine"), Value: &wm.Name},
		{Key: fn.Ptr("kloudlite.io/owner"), Value: &wm.Spec.OwnedBy},
		{Key: fn.Ptr("kloudlite.io/managed-by"), Value: fn.Ptr("kloudlite-controller")},
	}

	amiInfo, err := p.ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		ImageIds: []string{wm.Spec.AWSProvider.AMI},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get AMI info: %w", err)
	}

	userData, err := templates.K3sAgentSetup.Render(templates.K3sAgentSetupArgs{
		K3sURL:       p.k3sURL,
		K3sToken:     p.k3sToken,
		MachineName:  wm.Name,
		MachineOwner: wm.Spec.OwnedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to render k3s user data script: %w", err)
	}

	// Step 6: Build RunInstances input
	runInput := &ec2.RunInstancesInput{
		ImageId:          fn.Ptr(wm.Spec.AWSProvider.AMI),
		InstanceType:     wm.Spec.AWSProvider.MachineType,
		MinCount:         fn.Ptr[int32](1),
		MaxCount:         fn.Ptr[int32](1),
		SecurityGroupIds: securityGroupIDs,
		UserData:         fn.Ptr(base64.StdEncoding.EncodeToString(userData)),
		KeyName:          sshKeyPair,
		BlockDeviceMappings: []ec2types.BlockDeviceMapping{
			{
				DeviceName: amiInfo.Images[0].RootDeviceName,
				Ebs: &ec2types.EbsBlockDevice{
					VolumeSize:          &wm.Spec.AWSProvider.VolumeSize,
					VolumeType:          wm.Spec.AWSProvider.VolumeType,
					DeleteOnTermination: &wm.Spec.AWSProvider.DeleteVolumePostTermination,
				},
			},
		},
		TagSpecifications: []ec2types.TagSpecification{
			{ResourceType: ec2types.ResourceTypeInstance, Tags: tags},
			{ResourceType: ec2types.ResourceTypeVolume, Tags: tags},
		},
	}

	if wm.Spec.AWSProvider.IAMRole != nil {
		runInput.IamInstanceProfile = &ec2types.IamInstanceProfileSpecification{
			Name: wm.Spec.AWSProvider.IAMRole,
		}
	}

	// Step 7: Create instance
	runOutput, err := p.ec2Client.RunInstances(ctx, runInput)
	if err != nil {
		return nil, errors.Wrap("failed to create instance", err)
	}

	if len(runOutput.Instances) == 0 {
		return nil, errors.New("no instances were created")
	}

	instance := runOutput.Instances[0]

	return &v1.MachineInfo{
		MachineID:        aws.ToString(instance.InstanceId),
		State:            mapEC2StateToMachineState(instance.State),
		PrivateIP:        aws.ToString(instance.PrivateIpAddress),
		PublicIP:         aws.ToString(instance.PublicIpAddress),
		AvailabilityZone: aws.ToString(instance.Placement.AvailabilityZone),
		Message:          "Instance created successfully",
	}, nil
}

func (p *provider) GetMachineStatus(ctx context.Context, machineID string) (*v1.MachineInfo, error) {
	if machineID == "" {
		return nil, errors.Wrap("must provide machineID")
	}

	output, err := p.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{machineID},
	})
	if err != nil {
		return nil, errors.Wrap("failed to find machine", err)
	}

	if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return nil, errors.New("failed to find machine")
	}

	instance := output.Reservations[0].Instances[0]
	return &v1.MachineInfo{
		MachineID:        aws.ToString(instance.InstanceId),
		State:            mapEC2StateToMachineState(instance.State),
		PrivateIP:        aws.ToString(instance.PrivateIpAddress),
		PublicIP:         aws.ToString(instance.PublicIpAddress),
		AvailabilityZone: aws.ToString(instance.Placement.AvailabilityZone),
		Message:          fmt.Sprintf("Instance is %s", string(instance.State.Name)),
	}, nil
}

func (p *provider) StartMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	if _, err := p.ec2Client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{machineID},
	}); err != nil {
		return fmt.Errorf("failed to start machine: %w", err)
	}
	return nil
}

func (p *provider) DeleteMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	if _, err := p.ec2Client.TerminateInstances(ctx, &ec2.TerminateInstancesInput{
		InstanceIds: []string{machineID},
	}); err != nil {
		return fmt.Errorf("failed to delete machine: %w", err)
	}
	return nil
}

func (p *provider) IncreaseVolumeSize(ctx context.Context, machineID string, newSize int32) error {
	if machineID == "" || newSize == 0 {
		return errors.New("must provide machineID and newSize")
	}

	// Get instance to find root volume
	output, err := p.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{machineID},
	})
	if err != nil {
		return errors.Wrap("failed to find machine", err)
	}

	if len(output.Reservations) != 1 {
		return errors.New("failed to find machine")
	}

	instance := output.Reservations[0].Instances[0]

	var volumeID string
	for _, mapping := range instance.BlockDeviceMappings {
		if aws.ToString(mapping.DeviceName) == *instance.RootDeviceName {
			volumeID = aws.ToString(mapping.Ebs.VolumeId)
			break
		}
	}
	if volumeID == "" {
		return fmt.Errorf("root volume not found")
	}

	// Get current size
	volOutput, err := p.ec2Client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
		VolumeIds: []string{volumeID},
	})
	if err != nil || len(volOutput.Volumes) == 0 {
		return fmt.Errorf("volume not found: %s", volumeID)
	}

	currentSize := aws.ToInt32(volOutput.Volumes[0].Size)
	if newSize <= currentSize {
		return fmt.Errorf("new size (%d GB) must be greater than current size (%d GB)", newSize, currentSize)
	}

	_, err = p.ec2Client.ModifyVolume(ctx, &ec2.ModifyVolumeInput{
		VolumeId: aws.String(volumeID),
		Size:     aws.Int32(newSize),
	})
	return err
}

func (p *provider) ChangeMachine(ctx context.Context, machineID string, newInstanceType string) error {
	if machineID == "" || newInstanceType == "" {
		return errors.New("must provide machineID and newInstanceType")
	}

	// Stop instance
	if _, err := p.ec2Client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{machineID},
	}); err != nil {
		return fmt.Errorf("failed to stop instance: %w", err)
	}

	// Wait for stopped
	waiter := ec2.NewInstanceStoppedWaiter(p.ec2Client)
	if err := waiter.Wait(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{machineID},
	}, 300); err != nil {
		return fmt.Errorf("failed waiting for instance to stop: %w", err)
	}

	// Modify instance type
	if _, err := p.ec2Client.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
		InstanceId:   aws.String(machineID),
		InstanceType: &ec2types.AttributeValue{Value: aws.String(newInstanceType)},
	}); err != nil {
		return fmt.Errorf("failed to modify instance type: %w", err)
	}

	// Start instance
	if _, err := p.ec2Client.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: []string{machineID},
	}); err != nil {
		return fmt.Errorf("failed to start instance: %w", err)
	}

	return nil
}

func mapEC2StateToMachineState(state *ec2types.InstanceState) v1.MachineState {
	if state == nil {
		return v1.MachineStateError
	}
	switch state.Name {
	case ec2types.InstanceStateNamePending:
		return v1.MachineStateStarting
	case ec2types.InstanceStateNameRunning:
		return v1.MachineStateRunning
	case ec2types.InstanceStateNameStopping:
		return v1.MachineStateStopping
	case ec2types.InstanceStateNameStopped:
		return v1.MachineStateStopped
	default:
		return v1.MachineStateError
	}
}

// Helper methods

func (p *provider) getOrCreateSecurityGroup(
	ctx context.Context, vpcID string, allowedCIDR string,
) (string, error) {
	name := "kloudlite-workmachines"

	// Try to find existing security group
	out, err := p.ec2Client.DescribeSecurityGroups(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupNames: []string{name},
		Filters: []ec2types.Filter{
			{Name: fn.Ptr("group-name"), Values: []string{name}},
			{Name: fn.Ptr("vpc-id"), Values: []string{vpcID}},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get security groups: %w", err)
	}

	if len(out.SecurityGroups) == 1 {
		return aws.ToString(out.SecurityGroups[0].GroupId), nil
	}

	// Create new security group
	sg, err := p.ec2Client.CreateSecurityGroup(ctx, &ec2.CreateSecurityGroupInput{
		GroupName:   fn.Ptr(name),
		Description: fn.Ptr(fmt.Sprintf("Security group for kloudlite workmachines")),
		VpcId:       fn.Ptr(vpcID),
		TagSpecifications: []ec2types.TagSpecification{
			{
				ResourceType: ec2types.ResourceTypeSecurityGroup,
				Tags: []ec2types.Tag{
					{Key: fn.Ptr("Name"), Value: fn.Ptr(name)},
					{Key: fn.Ptr("kloudlite.io/managed-by"), Value: fn.Ptr("kloudlite-controller")},
				},
			},
		},
	})
	if err != nil {
		return "", errors.Wrap("failed to create security group", err)
	}

	sgID := sg.GroupId

	// Add ingress rules for HTTPS (443) and SSH (22)
	_, err = p.ec2Client.AuthorizeSecurityGroupIngress(ctx, &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: sgID,
		IpPermissions: []ec2types.IpPermission{
			{
				IpProtocol: fn.Ptr("tcp"),
				FromPort:   fn.Ptr[int32](443),
				ToPort:     fn.Ptr[int32](443),
				IpRanges:   []ec2types.IpRange{{CidrIp: &allowedCIDR, Description: fn.Ptr("HTTPS access")}},
			},
		},
	})
	if err != nil && !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
		return "", errors.Wrap("failed to authorize security group ingress", err)
	}

	// Add egress rule (allow all outbound)
	_, err = p.ec2Client.AuthorizeSecurityGroupEgress(ctx, &ec2.AuthorizeSecurityGroupEgressInput{
		GroupId: sgID,
		IpPermissions: []ec2types.IpPermission{
			{
				IpProtocol: fn.Ptr("-1"),
				IpRanges:   []ec2types.IpRange{{CidrIp: fn.Ptr("0.0.0.0/0"), Description: fn.Ptr("Allow all outbound")}},
			},
		},
	})
	if err != nil && !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
		return "", errors.Wrap("failed to authorize security group egress", err)
	}

	return *sgID, nil
}

func (p *provider) getOrCreateSSHKeyPair(ctx context.Context) (*string, error) {
	name := "kloudlite-workmachine"
	result, err := p.ec2Client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
		KeyNames: []string{name},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe AWS key pairs: %w", err)
	}

	if len(result.KeyPairs) == 1 {
		return result.KeyPairs[0].KeyName, nil
	}

	slog.Info("kloudlite workmachine ssh secret not found, will be creating a new one")
	out, err := p.ec2Client.CreateKeyPair(ctx, &ec2.CreateKeyPairInput{KeyName: &name})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS key pair: %w", err)
	}

	return out.KeyName, nil
}
