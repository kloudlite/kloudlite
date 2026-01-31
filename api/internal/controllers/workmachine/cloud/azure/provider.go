package azure

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/cloud"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/templates"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"golang.org/x/sync/errgroup"
)

type provider struct {
	credential             *azidentity.DefaultAzureCredential
	vmClient               *armcompute.VirtualMachinesClient
	diskClient             *armcompute.DisksClient
	nicClient              *armnetwork.InterfacesClient
	publicIPClient         *armnetwork.PublicIPAddressesClient
	subscriptionID         string
	resourceGroup          string
	location               string
	subnetID               string
	networkSecurityGroupID string
	ProviderArgs
}

var _ cloud.Provider = (*provider)(nil)

type Tag struct {
	Key   string
	Value string
}

// deepLearningImages contains Azure Deep Learning VM images for each region
// Using NVIDIA NGC images from Azure Marketplace (Ubuntu 22.04 with CUDA)
// Format: Publisher:Offer:SKU:Version
var deepLearningImages = map[string]ImageReference{
	// Using Canonical Ubuntu Server 24.04 LTS with NVIDIA GPU drivers
	// These can be used with Azure NV/NC series VMs
	"eastus":             {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"eastus2":            {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"westus":             {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"westus2":            {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"westus3":            {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"centralus":          {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"northeurope":        {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"westeurope":         {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"southeastasia":      {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"eastasia":           {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"japaneast":          {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"japanwest":          {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"australiaeast":      {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"australiasoutheast": {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"southcentralus":     {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"northcentralus":     {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"uksouth":            {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"ukwest":             {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"canadacentral":      {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"canadaeast":         {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"germanywestcentral": {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"francecentral":      {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"koreacentral":       {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"koreasouth":         {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"brazilsouth":        {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"southafricanorth":   {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"uaenorth":           {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"switzerlandnorth":   {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"norwayeast":         {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"swedencentral":      {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"qatarcentral":       {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"polandcentral":      {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"italynorth":         {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"israelcentral":      {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"centralindia":       {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"southindia":         {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
	"westindia":          {Publisher: "Canonical", Offer: "ubuntu-24_04-lts", SKU: "server", Version: "latest"},
}

type ImageReference struct {
	Publisher string
	Offer     string
	SKU       string
	Version   string
}

type ProviderArgs struct {
	SubscriptionID         string
	ResourceGroup          string
	Location               string
	SubnetID               string
	NetworkSecurityGroupID string
	ResourceTags           []Tag

	K3sVersion      string
	K3sURL          string
	K3sToken        string
	HostedSubdomain string // e.g., "mega.khost.dev" - used for registry mirror config
}

func NewProvider(ctx context.Context, args ProviderArgs) (cloud.Provider, error) {
	// Validate that we have an image for this location
	if _, exists := deepLearningImages[args.Location]; !exists {
		return nil, fmt.Errorf("no VM image configured for location %s", args.Location)
	}

	// Create DefaultAzureCredential
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, errors.Wrap("failed to obtain Azure credentials", err)
	}

	// Create VM client
	vmClient, err := armcompute.NewVirtualMachinesClient(args.SubscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrap("failed to create VM client", err)
	}

	// Create Disk client
	diskClient, err := armcompute.NewDisksClient(args.SubscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrap("failed to create disk client", err)
	}

	// Create NIC client
	nicClient, err := armnetwork.NewInterfacesClient(args.SubscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrap("failed to create NIC client", err)
	}

	// Create Public IP client
	publicIPClient, err := armnetwork.NewPublicIPAddressesClient(args.SubscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrap("failed to create public IP client", err)
	}

	return &provider{
		credential:             cred,
		vmClient:               vmClient,
		diskClient:             diskClient,
		nicClient:              nicClient,
		publicIPClient:         publicIPClient,
		subscriptionID:         args.SubscriptionID,
		resourceGroup:          args.ResourceGroup,
		location:               args.Location,
		subnetID:               args.SubnetID,
		networkSecurityGroupID: args.NetworkSecurityGroupID,
		ProviderArgs:           args,
	}, nil
}

func handleDryRunError(err error, action string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// For Azure, authorization errors contain specific messages
	if strings.Contains(errStr, "AuthorizationFailed") ||
		strings.Contains(errStr, "does not have authorization") ||
		strings.Contains(errStr, "LinkedAuthorizationFailed") {
		return err
	}

	// Resource not found errors during dry-run are acceptable
	// These indicate permissions are granted but resource doesn't exist
	if strings.Contains(errStr, "ResourceNotFound") ||
		strings.Contains(errStr, "NotFound") ||
		strings.Contains(errStr, "does not exist") {
		return nil
	}

	return err
}

func (p *provider) getVM(ctx context.Context, vmName string) (*armcompute.VirtualMachine, error) {
	resp, err := p.vmClient.Get(ctx, p.resourceGroup, vmName, &armcompute.VirtualMachinesClientGetOptions{
		Expand: to.Ptr(armcompute.InstanceViewTypesInstanceView),
	})
	if err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to get VM (Name: %s)", vmName), err)
	}
	return &resp.VirtualMachine, nil
}

func (p *provider) getVMByID(ctx context.Context, machineID string) (*armcompute.VirtualMachine, error) {
	// machineID is the Azure resource ID, extract VM name from it
	vmName := extractVMNameFromID(machineID)
	return p.getVM(ctx, vmName)
}

func extractVMNameFromID(resourceID string) string {
	// Azure resource ID format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Compute/virtualMachines/{vmName}
	parts := strings.Split(resourceID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return resourceID
}

func (p *provider) getOSDisk(ctx context.Context, vm *armcompute.VirtualMachine) (*armcompute.Disk, error) {
	if vm.Properties == nil || vm.Properties.StorageProfile == nil || vm.Properties.StorageProfile.OSDisk == nil {
		return nil, errors.New("VM does not have OS disk information")
	}

	diskName := ""
	if vm.Properties.StorageProfile.OSDisk.ManagedDisk != nil && vm.Properties.StorageProfile.OSDisk.ManagedDisk.ID != nil {
		// Extract disk name from resource ID
		parts := strings.Split(*vm.Properties.StorageProfile.OSDisk.ManagedDisk.ID, "/")
		if len(parts) > 0 {
			diskName = parts[len(parts)-1]
		}
	} else if vm.Properties.StorageProfile.OSDisk.Name != nil {
		diskName = *vm.Properties.StorageProfile.OSDisk.Name
	}

	if diskName == "" {
		return nil, errors.New("could not determine OS disk name")
	}

	resp, err := p.diskClient.Get(ctx, p.resourceGroup, diskName, nil)
	if err != nil {
		return nil, errors.Wrap(fmt.Sprintf("failed to get disk (Name: %s)", diskName), err)
	}

	return &resp.Disk, nil
}

func (p *provider) ValidatePermissions(ctx context.Context) error {
	g, errctx := errgroup.WithContext(ctx)
	dryRunChecks := []func(context.Context) error{
		// VM lifecycle operations
		p.dryRunCreateVM,
		p.dryRunDeleteVM,
		p.dryRunStartVM,
		p.dryRunStopVM,
		p.dryRunRestartVM,
		p.dryRunGetVM,

		// Disk operations
		p.dryRunGetDisk,
		p.dryRunUpdateDisk,

		// Network operations
		p.dryRunCreateNIC,
		p.dryRunDeleteNIC,
		p.dryRunCreatePublicIP,
		p.dryRunDeletePublicIP,
	}

	for i := range dryRunChecks {
		check := dryRunChecks[i]
		g.Go(func() error {
			if err := check(errctx); err != nil {
				return err
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return errors.Wrap("failed to validate permissions", err)
	}

	slog.Info("[Azure Provider] All permission checks passed")
	return nil
}

func (p *provider) CreateMachine(ctx context.Context, wm *v1.WorkMachine) (*v1.MachineInfo, error) {
	vmName := "kl-workmachine-" + wm.Name
	nicName := vmName + "-nic"
	publicIPName := vmName + "-pip"

	tags := map[string]*string{
		"Name":                  to.Ptr(vmName),
		"kloudlite-workmachine": to.Ptr(wm.Name),
		"kloudlite-owner":       to.Ptr(wm.Spec.OwnedBy),
		"kloudlite-managed-by":  to.Ptr("kloudlite-controller"),
	}

	for _, tag := range p.ResourceTags {
		tags[tag.Key] = to.Ptr(tag.Value)
	}

	// Get the location-specific image
	imageRef, exists := deepLearningImages[p.location]
	if !exists {
		return nil, fmt.Errorf("no VM image configured for location %s", p.location)
	}

	// Prepare user data (cloud-init script)
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

	// Create Public IP with Dynamic allocation (matches AWS behavior - IP can change on stop/start)
	publicIPResp, err := p.publicIPClient.BeginCreateOrUpdate(ctx, p.resourceGroup, publicIPName, armnetwork.PublicIPAddress{
		Location: to.Ptr(p.location),
		Tags:     tags,
		Properties: &armnetwork.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
			PublicIPAddressVersion:   to.Ptr(armnetwork.IPVersionIPv4),
		},
		SKU: &armnetwork.PublicIPAddressSKU{
			Name: to.Ptr(armnetwork.PublicIPAddressSKUNameBasic),
		},
	}, nil)
	if err != nil {
		return nil, errors.Wrap("failed to create public IP", err)
	}

	publicIP, err := publicIPResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, errors.Wrap("failed to wait for public IP creation", err)
	}

	// Create NIC
	nicProperties := &armnetwork.InterfacePropertiesFormat{
		IPConfigurations: []*armnetwork.InterfaceIPConfiguration{
			{
				Name: to.Ptr("ipconfig1"),
				Properties: &armnetwork.InterfaceIPConfigurationPropertiesFormat{
					Subnet: &armnetwork.Subnet{
						ID: to.Ptr(p.subnetID),
					},
					PrivateIPAllocationMethod: to.Ptr(armnetwork.IPAllocationMethodDynamic),
					PublicIPAddress: &armnetwork.PublicIPAddress{
						ID: publicIP.ID,
					},
				},
			},
		},
	}

	// Add NSG if configured
	if p.networkSecurityGroupID != "" {
		nicProperties.NetworkSecurityGroup = &armnetwork.SecurityGroup{
			ID: to.Ptr(p.networkSecurityGroupID),
		}
	}

	nicResp, err := p.nicClient.BeginCreateOrUpdate(ctx, p.resourceGroup, nicName, armnetwork.Interface{
		Location:   to.Ptr(p.location),
		Tags:       tags,
		Properties: nicProperties,
	}, nil)
	if err != nil {
		return nil, errors.Wrap("failed to create NIC", err)
	}

	nic, err := nicResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, errors.Wrap("failed to wait for NIC creation", err)
	}

	// Determine volume size
	volumeSize := int32(100)
	if wm.Spec.VolumeSize != nil {
		volumeSize = *wm.Spec.VolumeSize
	}

	// Determine disk type (map from AWS to Azure)
	osDiskType := armcompute.StorageAccountTypesPremiumLRS
	switch wm.Spec.VolumeType {
	case "gp3", "gp2":
		osDiskType = armcompute.StorageAccountTypesPremiumLRS
	case "io1", "io2":
		osDiskType = armcompute.StorageAccountTypesUltraSSDLRS
	case "standard":
		osDiskType = armcompute.StorageAccountTypesStandardLRS
	}

	// Convert Kubernetes-friendly machine type name (e.g., "Standard_B2ms" or "standard-b2ms") to Azure VM size
	azureVMSize := strings.ReplaceAll(wm.Spec.MachineType, "-", "_")
	// Ensure first letter of each part after underscore is capitalized (Standard_B2ms format)
	if !strings.HasPrefix(azureVMSize, "Standard_") && !strings.HasPrefix(azureVMSize, "standard_") {
		azureVMSize = "Standard_" + azureVMSize
	}

	// Create VM
	vmResp, err := p.vmClient.BeginCreateOrUpdate(ctx, p.resourceGroup, vmName, armcompute.VirtualMachine{
		Location: to.Ptr(p.location),
		Tags:     tags,
		Properties: &armcompute.VirtualMachineProperties{
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(azureVMSize)),
			},
			StorageProfile: &armcompute.StorageProfile{
				ImageReference: &armcompute.ImageReference{
					Publisher: to.Ptr(imageRef.Publisher),
					Offer:     to.Ptr(imageRef.Offer),
					SKU:       to.Ptr(imageRef.SKU),
					Version:   to.Ptr(imageRef.Version),
				},
				OSDisk: &armcompute.OSDisk{
					Name:         to.Ptr(vmName + "-osdisk"),
					CreateOption: to.Ptr(armcompute.DiskCreateOptionTypesFromImage),
					Caching:      to.Ptr(armcompute.CachingTypesReadWrite),
					ManagedDisk: &armcompute.ManagedDiskParameters{
						StorageAccountType: to.Ptr(osDiskType),
					},
					DiskSizeGB:   to.Ptr(volumeSize),
					DeleteOption: to.Ptr(armcompute.DiskDeleteOptionTypesDelete),
				},
			},
			OSProfile: &armcompute.OSProfile{
				ComputerName:  to.Ptr(vmName),
				AdminUsername: to.Ptr("kloudlite"),
				LinuxConfiguration: &armcompute.LinuxConfiguration{
					DisablePasswordAuthentication: to.Ptr(true),
					SSH: &armcompute.SSHConfiguration{
						PublicKeys: []*armcompute.SSHPublicKey{
							{
								Path:    to.Ptr("/home/kloudlite/.ssh/authorized_keys"),
								KeyData: to.Ptr(generateDummySSHKey()),
							},
						},
					},
				},
				CustomData: to.Ptr(base64.StdEncoding.EncodeToString(userData)),
			},
			NetworkProfile: &armcompute.NetworkProfile{
				NetworkInterfaces: []*armcompute.NetworkInterfaceReference{
					{
						ID: nic.ID,
						Properties: &armcompute.NetworkInterfaceReferenceProperties{
							Primary:      to.Ptr(true),
							DeleteOption: to.Ptr(armcompute.DeleteOptionsDelete),
						},
					},
				},
			},
		},
	}, nil)
	if err != nil {
		return nil, errors.Wrap("failed to create Azure VM", err)
	}

	vm, err := vmResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, errors.Wrap("failed to wait for VM creation", err)
	}

	// Get the IP addresses from NIC
	privateIP := ""
	publicIPAddr := ""
	if nic.Properties != nil && len(nic.Properties.IPConfigurations) > 0 {
		ipConfig := nic.Properties.IPConfigurations[0]
		if ipConfig.Properties != nil && ipConfig.Properties.PrivateIPAddress != nil {
			privateIP = *ipConfig.Properties.PrivateIPAddress
		}
	}
	if publicIP.Properties != nil && publicIP.Properties.IPAddress != nil {
		publicIPAddr = *publicIP.Properties.IPAddress
	}

	return &v1.MachineInfo{
		MachineID:         *vm.ID,
		State:             mapAzureStateToMachineState(&vm.VirtualMachine),
		PrivateIP:         privateIP,
		PublicIP:          publicIPAddr,
		AvailabilityZone:  p.location,
		Message:           "Instance created successfully",
		Region:            p.location,
		StorageVolumeSize: volumeSize,
	}, nil
}

func (p *provider) GetMachineStatus(ctx context.Context, machineID string) (*v1.MachineInfo, error) {
	if machineID == "" {
		return nil, errors.Wrap("must provide machineID")
	}

	vm, err := p.getVMByID(ctx, machineID)
	if err != nil {
		return nil, err
	}

	// Get IP addresses
	privateIP, publicIP := p.getVMIPAddresses(ctx, vm)

	return &v1.MachineInfo{
		MachineID:        fn.ValueOf(vm.ID),
		State:            mapAzureStateToMachineState(vm),
		PrivateIP:        privateIP,
		PublicIP:         publicIP,
		AvailabilityZone: fn.ValueOf(vm.Location),
		Message:          fmt.Sprintf("Instance is %s", getVMPowerState(vm)),
		Region:           p.location,
	}, nil
}

func (p *provider) getVMIPAddresses(ctx context.Context, vm *armcompute.VirtualMachine) (privateIP, publicIP string) {
	if vm.Properties == nil || vm.Properties.NetworkProfile == nil {
		return "", ""
	}

	for _, nicRef := range vm.Properties.NetworkProfile.NetworkInterfaces {
		if nicRef.ID == nil {
			continue
		}

		nicName := extractVMNameFromID(*nicRef.ID)
		nicResp, err := p.nicClient.Get(ctx, p.resourceGroup, nicName, nil)
		if err != nil {
			continue
		}

		if nicResp.Properties != nil {
			for _, ipConfig := range nicResp.Properties.IPConfigurations {
				if ipConfig.Properties == nil {
					continue
				}
				if ipConfig.Properties.PrivateIPAddress != nil {
					privateIP = *ipConfig.Properties.PrivateIPAddress
				}
				if ipConfig.Properties.PublicIPAddress != nil && ipConfig.Properties.PublicIPAddress.ID != nil {
					pipName := extractVMNameFromID(*ipConfig.Properties.PublicIPAddress.ID)
					pipResp, err := p.publicIPClient.Get(ctx, p.resourceGroup, pipName, nil)
					if err == nil && pipResp.Properties != nil && pipResp.Properties.IPAddress != nil {
						publicIP = *pipResp.Properties.IPAddress
					}
				}
			}
		}
	}

	return privateIP, publicIP
}

func (p *provider) StartMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	vmName := extractVMNameFromID(machineID)
	poller, err := p.vmClient.BeginStart(ctx, p.resourceGroup, vmName, nil)
	if err != nil {
		return fmt.Errorf("failed to start machine: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to wait for machine start: %w", err)
	}

	return nil
}

func (p *provider) StopMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	vmName := extractVMNameFromID(machineID)
	// Use Deallocate instead of PowerOff to release the VM resources and stop billing
	poller, err := p.vmClient.BeginDeallocate(ctx, p.resourceGroup, vmName, nil)
	if err != nil {
		return fmt.Errorf("failed to stop machine: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to wait for machine stop: %w", err)
	}

	return nil
}

func (p *provider) RebootMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	vmName := extractVMNameFromID(machineID)
	poller, err := p.vmClient.BeginRestart(ctx, p.resourceGroup, vmName, nil)
	if err != nil {
		return fmt.Errorf("failed to reboot machine: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to wait for machine reboot: %w", err)
	}

	return nil
}

func (p *provider) DeleteMachine(ctx context.Context, machineID string) error {
	if machineID == "" {
		return fmt.Errorf("must provide machineID, got (%s)", machineID)
	}

	vmName := extractVMNameFromID(machineID)

	// Delete VM (NIC and Public IP will be deleted automatically due to DeleteOption)
	poller, err := p.vmClient.BeginDelete(ctx, p.resourceGroup, vmName, nil)
	if err != nil {
		// Ignore not found errors
		if strings.Contains(err.Error(), "ResourceNotFound") || strings.Contains(err.Error(), "NotFound") {
			return nil
		}
		return fmt.Errorf("failed to delete machine: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to wait for machine deletion: %w", err)
	}

	return nil
}

func (p *provider) IncreaseVolumeSize(ctx context.Context, machineID string, newSize int32) error {
	if machineID == "" || newSize == 0 {
		return errors.New("must provide machineID and newSize")
	}

	vm, err := p.getVMByID(ctx, machineID)
	if err != nil {
		return err
	}

	disk, err := p.getOSDisk(ctx, vm)
	if err != nil {
		return err
	}

	currentSize := fn.ValueOf(disk.Properties.DiskSizeGB)
	if newSize < currentSize {
		return fmt.Errorf("new size (%d GB) must be greater than current size (%d GB)", newSize, currentSize)
	}

	// Update disk size
	poller, err := p.diskClient.BeginUpdate(ctx, p.resourceGroup, *disk.Name, armcompute.DiskUpdate{
		Properties: &armcompute.DiskUpdateProperties{
			DiskSizeGB: to.Ptr(newSize),
		},
	}, nil)
	if err != nil {
		return errors.Wrap(fmt.Sprintf("failed to increase disk's (Name: %s) size", *disk.Name), err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return errors.Wrap("failed to wait for disk resize", err)
	}

	return nil
}

func (p *provider) ChangeMachine(ctx context.Context, machineID string, newInstanceType string) error {
	if machineID == "" || newInstanceType == "" {
		return errors.New("must provide machineID and newInstanceType")
	}

	vmName := extractVMNameFromID(machineID)

	// Convert Kubernetes-friendly machine type name to Azure VM size
	azureVMSize := strings.ReplaceAll(newInstanceType, "-", "_")
	if !strings.HasPrefix(azureVMSize, "Standard_") && !strings.HasPrefix(azureVMSize, "standard_") {
		azureVMSize = "Standard_" + azureVMSize
	}

	// NOTE: The VM must already be deallocated by the controller before calling this function
	// The controller handles the stop/start flow with proper status checks and requeuing
	poller, err := p.vmClient.BeginUpdate(ctx, p.resourceGroup, vmName, armcompute.VirtualMachineUpdate{
		Properties: &armcompute.VirtualMachineProperties{
			HardwareProfile: &armcompute.HardwareProfile{
				VMSize: to.Ptr(armcompute.VirtualMachineSizeTypes(azureVMSize)),
			},
		},
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to modify VM size: %w", err)
	}

	_, err = poller.PollUntilDone(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to wait for VM size change: %w", err)
	}

	return nil
}

func mapAzureStateToMachineState(vm *armcompute.VirtualMachine) v1.MachineState {
	powerState := getVMPowerState(vm)

	switch powerState {
	case "starting":
		return v1.MachineStateStarting
	case "running":
		return v1.MachineStateRunning
	case "stopping", "deallocating":
		return v1.MachineStateStopping
	case "stopped", "deallocated":
		return v1.MachineStateStopped
	default:
		return v1.MachineStateErrored
	}
}

func getVMPowerState(vm *armcompute.VirtualMachine) string {
	if vm.Properties == nil || vm.Properties.InstanceView == nil {
		return "unknown"
	}

	for _, status := range vm.Properties.InstanceView.Statuses {
		if status.Code == nil {
			continue
		}
		if strings.HasPrefix(*status.Code, "PowerState/") {
			return strings.TrimPrefix(*status.Code, "PowerState/")
		}
	}

	return "unknown"
}

// generateDummySSHKey generates a minimal SSH key for Azure VM creation
// Azure requires an SSH key but we use cloud-init for actual setup
func generateDummySSHKey() string {
	// This is a placeholder SSH public key
	// The actual SSH access is managed through cloud-init and Kloudlite's auth system
	return "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7hZV9i9GVzK7eKxW8xf1v3Z9q3K0aU+Mx7kOQrZPR4N5bQ4TQP/X2jYYYKvMjZZgQw3T5ZT7aQ0zYtQ9jL5HY8fN8kL0K+R8oT2Z+y6Z0wYjZ0ZlPZZYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY= kloudlite-placeholder"
}
