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
	ProviderArgs
}

var _ cloud.Provider = (*provider)(nil)

type Tag struct {
	Key   string
	Value string
}

type ProviderArgs struct {
	AMI             string
	Region          string
	VPC             string
	SecurityGroupID string
	ResourceTags    []Tag

	K3sVersion string
	K3sURL     string
	K3sToken   string
}

func NewProvider(ctx context.Context, args ProviderArgs) (cloud.Provider, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(args.Region))
	if err != nil {
		return nil, errors.Wrap("failed to load AWS config", err)
	}

	ec2Client := ec2.NewFromConfig(awsCfg)

	return &provider{
		ec2Client:    ec2Client,
		ProviderArgs: args,
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

func (p *provider) getMachine(ctx context.Context, machineID string) (*ec2types.Instance, error) {
	output, err := p.ec2Client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{machineID},
	})
	if err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to find machine (ID: %s)", machineID), err)
	}

	if len(output.Reservations) == 0 || len(output.Reservations[0].Instances) == 0 {
		return nil, errors.New(fmt.Sprintf("failed to find machine (ID: %s)", machineID))
	}

	return &output.Reservations[0].Instances[0], nil
}

func (p *provider) getRootVolume(ctx context.Context, instance *ec2types.Instance) (*ec2types.Volume, error) {
	var volumeID *string
	for _, m := range instance.BlockDeviceMappings {
		if m.DeviceName == instance.RootDeviceName && m.Ebs != nil {
			volumeID = m.Ebs.VolumeId
		}
	}

	if volumeID == nil {
		return nil, errors.New("root volume not found")
	}

	// Get current size
	output, err := p.ec2Client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
		VolumeIds: []string{*volumeID},
	})
	if err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to get volume (ID: %s)", *volumeID), err)
	}

	if len(output.Volumes) == 0 {
		return nil, errors.New(fmt.Sprintf("volume (ID: %s) not found", *volumeID))
	}

	return &output.Volumes[0], nil
}

func (p *provider) ValidatePermissions(ctx context.Context) error {
	g, errctx := errgroup.WithContext(ctx)
	dryRunChecks := []func(context.Context) error{
		p.dryRunCreateInstance,
		p.dryRunTerminateInstance,
		// p.dryRunCreateSecurityGroup,
		// p.dryRunAuthorizeSecurityGroupIngress,
		// p.dryRunAuthorizeSecurityGroupEgress,
		// p.dryRunDeleteSecurityGroup,
		p.dryRunDescribeVolumes,
		p.dryRunDescribeInstanceStatus,
		// p.dryRunDescribeSecurityGroups,
		p.dryRunDescribeInstances,
	}

	for i := range dryRunChecks {
		g.Go(func() error {
			if err := dryRunChecks[i](errctx); err != nil {
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap("failed to validate permissions", err)
	}

	slog.Info("[AWS Provider] Dry Run check passed")
	return nil
}

func (p *provider) CreateMachine(ctx context.Context, wm *v1.WorkMachine) (*v1.MachineInfo, error) {
	tags := []ec2types.Tag{
		{Key: fn.Ptr("Name"), Value: fn.Ptr("kl-workmachine-" + wm.Name)},
		{Key: fn.Ptr("kloudlite.io/workmachine"), Value: &wm.Name},
		{Key: fn.Ptr("kloudlite.io/owner"), Value: &wm.Spec.OwnedBy},
		{Key: fn.Ptr("kloudlite.io/managed-by"), Value: fn.Ptr("kloudlite-controller")},
	}

	for _, tag := range p.ResourceTags {
		tags = append(tags, ec2types.Tag{Key: &tag.Key, Value: &tag.Value})
	}

	amiInfo, err := p.ec2Client.DescribeImages(ctx, &ec2.DescribeImagesInput{
		ImageIds: []string{p.AMI},
	})
	if err != nil {
		return nil, errors.Wrap("failed to get AMI info", err)
	}

	userData, err := templates.K3sAgentSetup.Render(templates.K3sAgentSetupArgs{
		K3sVersion:    p.K3sVersion,
		K3sURL:        p.K3sURL,
		K3sAgentToken: p.K3sToken,
		MachineName:   wm.Name,
		MachineOwner:  wm.Spec.OwnedBy,
	})
	if err != nil {
		return nil, errors.Wrap("failed to render k3s user data script", err)
	}

	volumeType := ec2types.VolumeType(wm.Spec.VolumeType)
	if volumeType == "" {
		volumeType = ec2types.VolumeTypeGp3
	}

	// Step 6: Build RunInstances input
	runInput := &ec2.RunInstancesInput{
		ImageId:          fn.Ptr(p.AMI),
		InstanceType:     ec2types.InstanceType(wm.Spec.MachineType),
		MinCount:         fn.Ptr[int32](1),
		MaxCount:         fn.Ptr[int32](1),
		SecurityGroupIds: []string{p.SecurityGroupID},
		UserData:         fn.Ptr(base64.StdEncoding.EncodeToString(userData)),
		BlockDeviceMappings: []ec2types.BlockDeviceMapping{
			{
				DeviceName: amiInfo.Images[0].RootDeviceName,
				Ebs: &ec2types.EbsBlockDevice{
					VolumeSize:          &wm.Spec.VolumeSize,
					VolumeType:          volumeType,
					DeleteOnTermination: &wm.Spec.DeleteVolumePostTermination,
				},
			},
		},
		TagSpecifications: []ec2types.TagSpecification{
			{ResourceType: ec2types.ResourceTypeInstance, Tags: tags},
			{ResourceType: ec2types.ResourceTypeVolume, Tags: tags},
		},
	}

	if wm.Spec.AWSProviderExtras != nil && wm.Spec.AWSProviderExtras.IAMRole != nil {
		runInput.IamInstanceProfile = &ec2types.IamInstanceProfileSpecification{
			Name: wm.Spec.AWSProviderExtras.IAMRole,
		}
	}

	// Step 7: Create instance
	runOutput, err := p.ec2Client.RunInstances(ctx, runInput)
	if err != nil {
		return nil, errors.Wrap("failed to create AWS instance", err)
	}

	if len(runOutput.Instances) == 0 {
		return nil, errors.New("no instances were created")
	}

	instance := runOutput.Instances[0]

	return &v1.MachineInfo{
		MachineID:        *instance.InstanceId,
		State:            mapEC2StateToMachineState(instance.State),
		PrivateIP:        fn.ValueOf(instance.PrivateIpAddress),
		PublicIP:         fn.ValueOf(instance.PublicIpAddress),
		AvailabilityZone: aws.ToString(instance.Placement.AvailabilityZone),
		Message:          "Instance created successfully",
		Region:           p.Region,
		RootVolumeSize:   wm.Spec.VolumeSize,
	}, nil
}

func (p *provider) GetMachineStatus(ctx context.Context, machineID string) (*v1.MachineInfo, error) {
	if machineID == "" {
		return nil, errors.Wrap("must provide machineID")
	}

	instance, err := p.getMachine(ctx, machineID)
	if err != nil {
		return nil, err
	}

	return &v1.MachineInfo{
		MachineID:        aws.ToString(instance.InstanceId),
		State:            mapEC2StateToMachineState(instance.State),
		PrivateIP:        aws.ToString(instance.PrivateIpAddress),
		PublicIP:         aws.ToString(instance.PublicIpAddress),
		AvailabilityZone: aws.ToString(instance.Placement.AvailabilityZone),
		Message:          fmt.Sprintf("Instance is %s", string(instance.State.Name)),
		Region:           p.Region,
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

func (p *provider) StopMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	if _, err := p.ec2Client.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: []string{machineID},
	}); err != nil {
		return fmt.Errorf("failed to start machine: %w", err)
	}
	return nil
}

func (p *provider) RebootMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	if _, err := p.ec2Client.RebootInstances(ctx, &ec2.RebootInstancesInput{
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

	instance, err := p.getMachine(ctx, machineID)
	if err != nil {
		return err
	}

	volume, err := p.getRootVolume(ctx, instance)

	currentSize := fn.ValueOf(volume.Size)
	if newSize < currentSize {
		return fmt.Errorf("new size (%d GB) must be greater than current size (%d GB)", newSize, currentSize)
	}

	if _, err = p.ec2Client.ModifyVolume(ctx, &ec2.ModifyVolumeInput{
		VolumeId: volume.VolumeId,
		Size:     &newSize,
	}); err != nil {
		return errors.Wrap(fmt.Sprintf("failed to increase volume's (ID: %s) size", *volume.VolumeId), err)
	}

	return nil
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
		return v1.MachineStateErrored
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
		return v1.MachineStateErrored
	}
}
