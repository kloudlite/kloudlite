package oci

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/templates"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
)

type provider struct {
	configProvider common.ConfigurationProvider
	computeClient  core.ComputeClient
	vnClient       core.VirtualNetworkClient

	ProviderArgs
}

var _ cloud.Provider = (*provider)(nil)

type Tag struct {
	Key   string
	Value string
}

type ProviderArgs struct {
	CompartmentID      string
	Region             string
	SubnetID           string
	NSGID              string
	AvailabilityDomain string

	ResourceTags []Tag

	K3sVersion      string
	K3sURL          string
	K3sToken        string
	HostedSubdomain string
}

func NewProvider(ctx context.Context, args ProviderArgs) (cloud.Provider, error) {
	// Use instance principal auth (running on OCI instance)
	configProvider, err := auth.InstancePrincipalConfigurationProvider()
	if err != nil {
		return nil, errors.Wrap("failed to create OCI instance principal config provider", err)
	}

	computeClient, err := core.NewComputeClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, errors.Wrap("failed to create OCI compute client", err)
	}

	vnClient, err := core.NewVirtualNetworkClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, errors.Wrap("failed to create OCI virtual network client", err)
	}

	// If availability domain is not set, discover it
	if args.AvailabilityDomain == "" {
		identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
		if err != nil {
			return nil, errors.Wrap("failed to create OCI identity client", err)
		}

		resp, err := identityClient.ListAvailabilityDomains(ctx, identity.ListAvailabilityDomainsRequest{
			CompartmentId: &args.CompartmentID,
		})
		if err != nil {
			return nil, errors.Wrap("failed to list availability domains", err)
		}

		if len(resp.Items) > 0 && resp.Items[0].Name != nil {
			args.AvailabilityDomain = *resp.Items[0].Name
		} else {
			return nil, errors.New("no availability domains found")
		}
	}

	return &provider{
		configProvider: configProvider,
		computeClient:  computeClient,
		vnClient:       vnClient,
		ProviderArgs:   args,
	}, nil
}

func (p *provider) ValidatePermissions(ctx context.Context) error {
	// Try listing instances to validate compute permissions
	resp, err := p.computeClient.ListInstances(ctx, core.ListInstancesRequest{
		CompartmentId: &p.CompartmentID,
		Limit:         fn.Ptr(1),
	})
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "401") || strings.Contains(errStr, "403") ||
			strings.Contains(errStr, "NotAuthenticated") || strings.Contains(errStr, "NotAuthorized") {
			return errors.Wrap("OCI compute permission check failed", err)
		}
		// Other errors (like empty results) are OK
	}
	_ = resp

	slog.Info("[OCI Provider] Permission checks passed")
	return nil
}

func (p *provider) CreateMachine(ctx context.Context, wm *v1.WorkMachine) (*v1.MachineInfo, error) {
	instanceName := "kl-workmachine-" + wm.Name

	// Parse shape and resources from machine type name
	shapeName, ocpus, memoryGB := parseMachineType(wm.Spec.MachineType)

	// Find Ubuntu 24.04 image for this shape
	imageID, err := p.findUbuntuImage(ctx, shapeName)
	if err != nil {
		return nil, errors.Wrap("failed to find Ubuntu image", err)
	}

	// Render user data script
	userData, err := templates.K3sAgentSetupOCI.Render(templates.K3sAgentSetupArgs{
		K3sVersion:      p.K3sVersion,
		K3sURL:          p.K3sURL,
		K3sAgentToken:   p.K3sToken,
		MachineName:     wm.Name,
		MachineOwner:    fn.LabelValueEncoder(wm.Spec.OwnedBy),
		HostedSubdomain: p.HostedSubdomain,
	})
	if err != nil {
		return nil, errors.Wrap("failed to render k3s user data script", err)
	}

	// Build freeform tags
	tags := map[string]string{
		"managed-by": "kloudlite",
		"project":    "kloudlite",
		"workmachine": wm.Name,
		"owner":       wm.Spec.OwnedBy,
	}
	for _, t := range p.ResourceTags {
		tags[t.Key] = t.Value
	}

	// Build metadata
	metadata := map[string]string{
		"user_data": base64.StdEncoding.EncodeToString(userData),
	}

	// Add SSH public keys if specified
	if len(wm.Spec.SSHPublicKeys) > 0 {
		metadata["ssh_authorized_keys"] = strings.Join(wm.Spec.SSHPublicKeys, "\n")
	}

	bootVolumeSizeInGBs := int64(fn.ValueOf(wm.Spec.VolumeSize))
	if bootVolumeSizeInGBs < 50 {
		bootVolumeSizeInGBs = 100
	}

	assignPublicIP := true
	launchResp, err := p.computeClient.LaunchInstance(ctx, core.LaunchInstanceRequest{
		LaunchInstanceDetails: core.LaunchInstanceDetails{
			AvailabilityDomain: &p.AvailabilityDomain,
			CompartmentId:      &p.CompartmentID,
			DisplayName:        &instanceName,
			Shape:              &shapeName,
			ShapeConfig: &core.LaunchInstanceShapeConfigDetails{
				Ocpus:       &ocpus,
				MemoryInGBs: &memoryGB,
			},
			SourceDetails: core.InstanceSourceViaImageDetails{
				ImageId:             &imageID,
				BootVolumeSizeInGBs: &bootVolumeSizeInGBs,
			},
			CreateVnicDetails: &core.CreateVnicDetails{
				SubnetId:       &p.SubnetID,
				NsgIds:         []string{p.NSGID},
				AssignPublicIp: &assignPublicIP,
			},
			Metadata:     metadata,
			FreeformTags: tags,
		},
	})
	if err != nil {
		return nil, errors.Wrap("failed to create OCI instance", err)
	}

	instanceID := ""
	if launchResp.Id != nil {
		instanceID = *launchResp.Id
	}

	return &v1.MachineInfo{
		MachineID:         instanceID,
		State:             v1.MachineStateStarting,
		AvailabilityZone:  p.AvailabilityDomain,
		Message:           "Instance created successfully",
		Region:            p.Region,
		StorageVolumeSize: int32(bootVolumeSizeInGBs),
	}, nil
}

func (p *provider) GetMachineStatus(ctx context.Context, machineID string) (*v1.MachineInfo, error) {
	if machineID == "" {
		return nil, errors.New("must provide machineID")
	}

	resp, err := p.computeClient.GetInstance(ctx, core.GetInstanceRequest{
		InstanceId: &machineID,
	})
	if err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to get instance %s", machineID), err)
	}

	publicIP, privateIP := p.getInstanceIPs(ctx, machineID)

	return &v1.MachineInfo{
		MachineID:        machineID,
		State:            mapOCIStateToMachineState(resp.LifecycleState),
		PrivateIP:        privateIP,
		PublicIP:         publicIP,
		AvailabilityZone: p.AvailabilityDomain,
		Message:          fmt.Sprintf("Instance is %s", resp.LifecycleState),
		Region:           p.Region,
	}, nil
}

func (p *provider) StartMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	action := core.InstanceActionActionStart
	_, err := p.computeClient.InstanceAction(ctx, core.InstanceActionRequest{
		InstanceId: &machineID,
		Action:     action,
	})
	if err != nil {
		return fmt.Errorf("failed to start machine: %w", err)
	}

	return nil
}

func (p *provider) StopMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	action := core.InstanceActionActionSoftstop
	_, err := p.computeClient.InstanceAction(ctx, core.InstanceActionRequest{
		InstanceId: &machineID,
		Action:     action,
	})
	if err != nil {
		return fmt.Errorf("failed to stop machine: %w", err)
	}

	return nil
}

func (p *provider) RebootMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	action := core.InstanceActionActionSoftreset
	_, err := p.computeClient.InstanceAction(ctx, core.InstanceActionRequest{
		InstanceId: &machineID,
		Action:     action,
	})
	if err != nil {
		return fmt.Errorf("failed to reboot machine: %w", err)
	}

	return nil
}

func (p *provider) IncreaseVolumeSize(ctx context.Context, machineID string, newSize int32) error {
	if machineID == "" || newSize == 0 {
		return errors.New("must provide machineID and newSize")
	}

	// Get boot volume attachment
	resp, err := p.computeClient.ListBootVolumeAttachments(ctx, core.ListBootVolumeAttachmentsRequest{
		AvailabilityDomain: &p.AvailabilityDomain,
		CompartmentId:      &p.CompartmentID,
		InstanceId:         &machineID,
	})
	if err != nil {
		return errors.Wrap("failed to list boot volume attachments", err)
	}

	if len(resp.Items) == 0 {
		return errors.New("no boot volume found for instance")
	}

	bootVolumeID := ""
	for _, att := range resp.Items {
		if att.BootVolumeId != nil && att.LifecycleState == core.BootVolumeAttachmentLifecycleStateAttached {
			bootVolumeID = *att.BootVolumeId
			break
		}
	}

	if bootVolumeID == "" {
		return errors.New("boot volume not found")
	}

	// Create block storage client
	blockClient, err := core.NewBlockstorageClientWithConfigurationProvider(p.configProvider)
	if err != nil {
		return errors.Wrap("failed to create block storage client", err)
	}

	newSizeGB := int64(newSize)
	_, err = blockClient.UpdateBootVolume(ctx, core.UpdateBootVolumeRequest{
		BootVolumeId: &bootVolumeID,
		UpdateBootVolumeDetails: core.UpdateBootVolumeDetails{
			SizeInGBs: &newSizeGB,
		},
	})
	if err != nil {
		return errors.Wrap("failed to resize boot volume", err)
	}

	return nil
}

func (p *provider) ChangeMachine(ctx context.Context, machineID string, newInstanceType string) error {
	if machineID == "" || newInstanceType == "" {
		return errors.New("must provide machineID and newInstanceType")
	}

	shapeName, ocpus, memoryGB := parseMachineType(newInstanceType)

	_, err := p.computeClient.UpdateInstance(ctx, core.UpdateInstanceRequest{
		InstanceId: &machineID,
		UpdateInstanceDetails: core.UpdateInstanceDetails{
			Shape: &shapeName,
			ShapeConfig: &core.UpdateInstanceShapeConfigDetails{
				Ocpus:       &ocpus,
				MemoryInGBs: &memoryGB,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to change machine type: %w", err)
	}

	return nil
}

func (p *provider) DeleteMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	preserveBootVolume := false
	_, err := p.computeClient.TerminateInstance(ctx, core.TerminateInstanceRequest{
		InstanceId:         &machineID,
		PreserveBootVolume: &preserveBootVolume,
	})
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete machine: %w", err)
	}

	return nil
}

// Helper functions

// parseMachineType converts a MachineType name like "vm-standard-e4-flex-1-8"
// to OCI shape name "VM.Standard.E4.Flex" with OCPU and memory values
func parseMachineType(machineType string) (shapeName string, ocpus float32, memoryGB float32) {
	// Default values
	shapeName = "VM.Standard.E4.Flex"
	ocpus = 1
	memoryGB = 8

	// Parse the machine type name: vm-standard-{series}-flex-{ocpus}-{memory}
	parts := strings.Split(machineType, "-")
	if len(parts) >= 6 {
		// Reconstruct shape name: VM.Standard.{Series}.Flex
		series := strings.ToUpper(parts[2]) // e4 -> E4
		shapeName = fmt.Sprintf("VM.Standard.%s.Flex", series)

		// Parse OCPU count
		if o, err := strconv.ParseFloat(parts[4], 32); err == nil {
			ocpus = float32(o)
		}

		// Parse memory GB
		if m, err := strconv.ParseFloat(parts[5], 32); err == nil {
			memoryGB = float32(m)
		}
	}

	return shapeName, ocpus, memoryGB
}

func mapOCIStateToMachineState(state core.InstanceLifecycleStateEnum) v1.MachineState {
	switch state {
	case core.InstanceLifecycleStateProvisioning, core.InstanceLifecycleStateStarting:
		return v1.MachineStateStarting
	case core.InstanceLifecycleStateRunning:
		return v1.MachineStateRunning
	case core.InstanceLifecycleStateStopping:
		return v1.MachineStateStopping
	case core.InstanceLifecycleStateStopped:
		return v1.MachineStateStopped
	case core.InstanceLifecycleStateTerminating, core.InstanceLifecycleStateTerminated:
		return v1.MachineStateStopped
	default:
		return v1.MachineStateErrored
	}
}

// getInstanceIPs returns the public and private IPs of an instance via VNIC lookup
func (p *provider) getInstanceIPs(ctx context.Context, instanceID string) (string, string) {
	resp, err := p.computeClient.ListVnicAttachments(ctx, core.ListVnicAttachmentsRequest{
		CompartmentId: &p.CompartmentID,
		InstanceId:    &instanceID,
	})
	if err != nil {
		return "", ""
	}

	for _, va := range resp.Items {
		if va.VnicId == nil || va.LifecycleState != core.VnicAttachmentLifecycleStateAttached {
			continue
		}

		vnicResp, err := p.vnClient.GetVnic(ctx, core.GetVnicRequest{
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

// findUbuntuImage finds the latest Ubuntu 24.04 amd64 image compatible with the given shape
func (p *provider) findUbuntuImage(ctx context.Context, shapeName string) (string, error) {
	resp, err := p.computeClient.ListImages(ctx, core.ListImagesRequest{
		CompartmentId:          &p.CompartmentID,
		OperatingSystem:        strPtr("Canonical Ubuntu"),
		OperatingSystemVersion: strPtr("24.04"),
		Shape:                  &shapeName,
		SortBy:                 core.ListImagesSortByTimecreated,
		SortOrder:              core.ListImagesSortOrderDesc,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list images: %w", err)
	}

	for _, img := range resp.Items {
		if img.Id == nil || img.DisplayName == nil {
			continue
		}
		if strings.Contains(strings.ToLower(*img.DisplayName), "aarch64") {
			continue
		}
		return *img.Id, nil
	}

	// Fallback without shape filter
	resp2, err := p.computeClient.ListImages(ctx, core.ListImagesRequest{
		CompartmentId:          &p.CompartmentID,
		OperatingSystem:        strPtr("Canonical Ubuntu"),
		OperatingSystemVersion: strPtr("24.04"),
		SortBy:                 core.ListImagesSortByTimecreated,
		SortOrder:              core.ListImagesSortOrderDesc,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list images (fallback): %w", err)
	}

	for _, img := range resp2.Items {
		if img.Id == nil || img.DisplayName == nil {
			continue
		}
		if strings.Contains(strings.ToLower(*img.DisplayName), "aarch64") {
			continue
		}
		return *img.Id, nil
	}

	return "", fmt.Errorf("no Ubuntu 24.04 image found")
}

func strPtr(s string) *string {
	return &s
}

