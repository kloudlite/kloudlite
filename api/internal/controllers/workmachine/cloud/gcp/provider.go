package gcp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/templates"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"golang.org/x/sync/errgroup"
)

type provider struct {
	instancesClient *compute.InstancesClient
	disksClient     *compute.DisksClient

	ProviderArgs
}

var _ cloud.Provider = (*provider)(nil)

type Tag struct {
	Key   string
	Value string
}

// Ubuntu 24.04 LTS image family - works in all regions
// Using image family allows GCP to automatically select the latest image
const ubuntu2404ImageFamily = "projects/ubuntu-os-cloud/global/images/family/ubuntu-2404-lts-amd64"

type ProviderArgs struct {
	Project    string
	Region     string
	Zone       string
	Network    string // VPC network name (e.g., "default")
	Subnetwork string // Subnet name (e.g., "default")

	ResourceTags []Tag

	K3sVersion      string
	K3sURL          string
	K3sToken        string
	HostedSubdomain string // e.g., "mega.khost.dev" - used for registry mirror config
}

func NewProvider(ctx context.Context, args ProviderArgs) (cloud.Provider, error) {
	instancesClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, errors.Wrap("failed to create GCP instances client", err)
	}

	disksClient, err := compute.NewDisksRESTClient(ctx)
	if err != nil {
		instancesClient.Close()
		return nil, errors.Wrap("failed to create GCP disks client", err)
	}

	return &provider{
		instancesClient: instancesClient,
		disksClient:     disksClient,
		ProviderArgs:    args,
	}, nil
}

func (p *provider) ValidatePermissions(ctx context.Context) error {
	g, errctx := errgroup.WithContext(ctx)
	permissionChecks := []func(context.Context) error{
		p.checkListInstances,
		p.checkListDisks,
	}

	for i := range permissionChecks {
		check := permissionChecks[i]
		g.Go(func() error {
			return check(errctx)
		})
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap("failed to validate GCP permissions", err)
	}

	slog.Info("[GCP Provider] All permission checks passed")
	return nil
}

func (p *provider) CreateMachine(ctx context.Context, wm *v1.WorkMachine) (*v1.MachineInfo, error) {
	instanceName := "kl-workmachine-" + wm.Name

	// Prepare labels (GCP uses labels instead of tags for metadata)
	labels := map[string]string{
		"kloudlite-workmachine": sanitizeLabel(wm.Name),
		"kloudlite-owner":       sanitizeLabel(wm.Spec.OwnedBy),
		"kloudlite-managed-by":  "kloudlite-controller",
	}
	for _, tag := range p.ResourceTags {
		labels[sanitizeLabel(tag.Key)] = sanitizeLabel(tag.Value)
	}

	// Render user data script
	userData, err := templates.K3sAgentSetup.Render(templates.K3sAgentSetupArgs{
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

	// Determine disk type
	diskType := getDiskType(wm.Spec.VolumeType, p.Project, p.Zone)

	// Network tags for firewall rules
	networkTags := []string{
		"kl-workmachine",
		fmt.Sprintf("kl-%s", wm.Name),
	}

	// Build instance configuration
	instance := &computepb.Instance{
		Name:        fn.Ptr(instanceName),
		MachineType: fn.Ptr(fmt.Sprintf("zones/%s/machineTypes/%s", p.Zone, wm.Spec.MachineType)),
		Disks: []*computepb.AttachedDisk{
			{
				AutoDelete: fn.Ptr(wm.Spec.DeleteVolumePostTermination),
				Boot:       fn.Ptr(true),
				InitializeParams: &computepb.AttachedDiskInitializeParams{
					SourceImage: fn.Ptr(ubuntu2404ImageFamily),
					DiskSizeGb:  fn.Ptr(int64(*wm.Spec.VolumeSize)),
					DiskType:    fn.Ptr(diskType),
					Labels:      labels,
				},
			},
		},
		NetworkInterfaces: []*computepb.NetworkInterface{
			{
				Network:    fn.Ptr(fmt.Sprintf("projects/%s/global/networks/%s", p.Project, p.Network)),
				Subnetwork: fn.Ptr(fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", p.Project, p.Region, p.Subnetwork)),
				AccessConfigs: []*computepb.AccessConfig{
					{
						Name:        fn.Ptr("External NAT"),
						Type:        fn.Ptr("ONE_TO_ONE_NAT"),
						NetworkTier: fn.Ptr("PREMIUM"),
					},
				},
			},
		},
		Metadata: &computepb.Metadata{
			Items: []*computepb.Items{
				{
					Key:   fn.Ptr("startup-script"),
					Value: fn.Ptr(string(userData)),
				},
			},
		},
		Labels: labels,
		Tags: &computepb.Tags{
			Items: networkTags,
		},
		Scheduling: &computepb.Scheduling{
			AutomaticRestart:  fn.Ptr(true),
			OnHostMaintenance: fn.Ptr("MIGRATE"),
		},
	}

	// Create the instance
	op, err := p.instancesClient.Insert(ctx, &computepb.InsertInstanceRequest{
		Project:          p.Project,
		Zone:             p.Zone,
		InstanceResource: instance,
	})
	if err != nil {
		return nil, errors.Wrap("failed to create GCP instance", err)
	}

	// Wait for operation to complete
	if err := op.Wait(ctx); err != nil {
		return nil, errors.Wrap("failed waiting for instance creation", err)
	}

	// Get the created instance details
	createdInstance, err := p.instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Project:  p.Project,
		Zone:     p.Zone,
		Instance: instanceName,
	})
	if err != nil {
		return nil, errors.Wrap("failed to get created instance details", err)
	}

	return &v1.MachineInfo{
		MachineID:        instanceName,
		State:            mapGCPStateToMachineState(createdInstance.Status),
		PrivateIP:        getPrivateIP(createdInstance),
		PublicIP:         getPublicIP(createdInstance),
		AvailabilityZone: p.Zone,
		Message:          "Instance created successfully",
		Region:           p.Region,
		StorageVolumeSize: *wm.Spec.VolumeSize,
	}, nil
}

func (p *provider) GetMachineStatus(ctx context.Context, machineID string) (*v1.MachineInfo, error) {
	if machineID == "" {
		return nil, errors.Wrap("must provide machineID")
	}

	instance, err := p.instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Project:  p.Project,
		Zone:     p.Zone,
		Instance: machineID,
	})
	if err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to get instance %s", machineID), err)
	}

	return &v1.MachineInfo{
		MachineID:        machineID,
		State:            mapGCPStateToMachineState(instance.Status),
		PrivateIP:        getPrivateIP(instance),
		PublicIP:         getPublicIP(instance),
		AvailabilityZone: p.Zone,
		Message:          fmt.Sprintf("Instance is %s", fn.ValueOf(instance.Status)),
		Region:           p.Region,
	}, nil
}

func (p *provider) StartMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	op, err := p.instancesClient.Start(ctx, &computepb.StartInstanceRequest{
		Project:  p.Project,
		Zone:     p.Zone,
		Instance: machineID,
	})
	if err != nil {
		return fmt.Errorf("failed to start machine: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for machine start: %w", err)
	}

	return nil
}

func (p *provider) StopMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	op, err := p.instancesClient.Stop(ctx, &computepb.StopInstanceRequest{
		Project:  p.Project,
		Zone:     p.Zone,
		Instance: machineID,
	})
	if err != nil {
		return fmt.Errorf("failed to stop machine: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for machine stop: %w", err)
	}

	return nil
}

func (p *provider) RebootMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	op, err := p.instancesClient.Reset(ctx, &computepb.ResetInstanceRequest{
		Project:  p.Project,
		Zone:     p.Zone,
		Instance: machineID,
	})
	if err != nil {
		return fmt.Errorf("failed to reboot machine: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for machine reboot: %w", err)
	}

	return nil
}

func (p *provider) DeleteMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	op, err := p.instancesClient.Delete(ctx, &computepb.DeleteInstanceRequest{
		Project:  p.Project,
		Zone:     p.Zone,
		Instance: machineID,
	})
	if err != nil {
		// If instance not found, consider it already deleted
		if strings.Contains(err.Error(), "notFound") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return fmt.Errorf("failed to delete machine: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for machine deletion: %w", err)
	}

	return nil
}

func (p *provider) IncreaseVolumeSize(ctx context.Context, machineID string, newSize int32) error {
	if machineID == "" || newSize == 0 {
		return errors.New("must provide machineID and newSize")
	}

	// Get instance to find boot disk
	instance, err := p.instancesClient.Get(ctx, &computepb.GetInstanceRequest{
		Project:  p.Project,
		Zone:     p.Zone,
		Instance: machineID,
	})
	if err != nil {
		return errors.Wrap("failed to get instance", err)
	}

	// Find the boot disk
	var bootDiskName string
	for _, disk := range instance.Disks {
		if disk.Boot != nil && *disk.Boot {
			// Extract disk name from source URL
			source := fn.ValueOf(disk.Source)
			parts := strings.Split(source, "/")
			if len(parts) > 0 {
				bootDiskName = parts[len(parts)-1]
			}
			break
		}
	}

	if bootDiskName == "" {
		return errors.New("boot disk not found")
	}

	// Resize the disk
	op, err := p.disksClient.Resize(ctx, &computepb.ResizeDiskRequest{
		Project: p.Project,
		Zone:    p.Zone,
		Disk:    bootDiskName,
		DisksResizeRequestResource: &computepb.DisksResizeRequest{
			SizeGb: fn.Ptr(int64(newSize)),
		},
	})
	if err != nil {
		return errors.Wrap("failed to resize disk", err)
	}

	if err := op.Wait(ctx); err != nil {
		return errors.Wrap("failed waiting for disk resize", err)
	}

	return nil
}

func (p *provider) ChangeMachine(ctx context.Context, machineID string, newInstanceType string) error {
	if machineID == "" || newInstanceType == "" {
		return errors.New("must provide machineID and newInstanceType")
	}

	// GCP machine type format is already correct (e.g., "n1-standard-4")
	// No conversion needed unlike AWS/Azure
	machineTypeURL := fmt.Sprintf("zones/%s/machineTypes/%s", p.Zone, newInstanceType)

	// NOTE: The instance must already be stopped by the controller before calling this function
	op, err := p.instancesClient.SetMachineType(ctx, &computepb.SetMachineTypeInstanceRequest{
		Project:  p.Project,
		Zone:     p.Zone,
		Instance: machineID,
		InstancesSetMachineTypeRequestResource: &computepb.InstancesSetMachineTypeRequest{
			MachineType: fn.Ptr(machineTypeURL),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to change machine type: %w", err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("failed waiting for machine type change: %w", err)
	}

	return nil
}

// Helper functions

func mapGCPStateToMachineState(status *string) v1.MachineState {
	if status == nil {
		return v1.MachineStateErrored
	}
	switch *status {
	case "PROVISIONING", "STAGING":
		return v1.MachineStateStarting
	case "RUNNING":
		return v1.MachineStateRunning
	case "STOPPING":
		return v1.MachineStateStopping
	case "TERMINATED", "STOPPED", "SUSPENDED":
		return v1.MachineStateStopped
	default:
		return v1.MachineStateErrored
	}
}

func getPrivateIP(instance *computepb.Instance) string {
	if instance.NetworkInterfaces != nil && len(instance.NetworkInterfaces) > 0 {
		return fn.ValueOf(instance.NetworkInterfaces[0].NetworkIP)
	}
	return ""
}

func getPublicIP(instance *computepb.Instance) string {
	if instance.NetworkInterfaces != nil && len(instance.NetworkInterfaces) > 0 {
		ni := instance.NetworkInterfaces[0]
		if ni.AccessConfigs != nil && len(ni.AccessConfigs) > 0 {
			return fn.ValueOf(ni.AccessConfigs[0].NatIP)
		}
	}
	return ""
}

func getDiskType(volumeType string, project, zone string) string {
	// Map common volume types to GCP disk types
	switch volumeType {
	case "ssd", "pd-ssd":
		return fmt.Sprintf("projects/%s/zones/%s/diskTypes/pd-ssd", project, zone)
	case "balanced", "pd-balanced":
		return fmt.Sprintf("projects/%s/zones/%s/diskTypes/pd-balanced", project, zone)
	case "standard", "pd-standard":
		return fmt.Sprintf("projects/%s/zones/%s/diskTypes/pd-standard", project, zone)
	default:
		// Default to balanced for good performance/cost ratio
		return fmt.Sprintf("projects/%s/zones/%s/diskTypes/pd-balanced", project, zone)
	}
}

// sanitizeLabel converts a string to a valid GCP label value
// GCP labels must be lowercase, start with a letter, and contain only lowercase letters, numbers, hyphens, and underscores
func sanitizeLabel(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, "@", "-at-")
	s = strings.ReplaceAll(s, "/", "-")

	// Ensure it starts with a letter
	if len(s) > 0 && (s[0] < 'a' || s[0] > 'z') {
		s = "l-" + s
	}

	// Truncate to max label length (63 chars)
	if len(s) > 63 {
		s = s[:63]
	}

	return s
}
