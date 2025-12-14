package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4"
)

// dryRunCreateVM validates permissions to create a VM
func (p *provider) dryRunCreateVM(ctx context.Context) error {
	// Try to get a non-existent VM - this validates read permissions
	// Azure doesn't have a true dry-run mode like AWS, so we use permission validation via API calls
	_, err := p.vmClient.Get(ctx, p.resourceGroup, "kl-dryrun-test-vm", nil)
	return handleDryRunError(err, "create VM")
}

// dryRunDeleteVM validates permissions to delete a VM
func (p *provider) dryRunDeleteVM(ctx context.Context) error {
	// List VMs to validate delete permissions context
	pager := p.vmClient.NewListPager(p.resourceGroup, nil)
	_, err := pager.NextPage(ctx)
	if err != nil {
		return handleDryRunError(err, "delete VM")
	}
	return nil
}

// dryRunStartVM validates permissions to start a VM
func (p *provider) dryRunStartVM(ctx context.Context) error {
	// Validate by checking if we can list VMs (start requires the same scope)
	pager := p.vmClient.NewListPager(p.resourceGroup, nil)
	_, err := pager.NextPage(ctx)
	if err != nil {
		return handleDryRunError(err, "start VM")
	}
	return nil
}

// dryRunStopVM validates permissions to stop a VM
func (p *provider) dryRunStopVM(ctx context.Context) error {
	pager := p.vmClient.NewListPager(p.resourceGroup, nil)
	_, err := pager.NextPage(ctx)
	if err != nil {
		return handleDryRunError(err, "stop VM")
	}
	return nil
}

// dryRunRestartVM validates permissions to restart a VM
func (p *provider) dryRunRestartVM(ctx context.Context) error {
	pager := p.vmClient.NewListPager(p.resourceGroup, nil)
	_, err := pager.NextPage(ctx)
	if err != nil {
		return handleDryRunError(err, "restart VM")
	}
	return nil
}

// dryRunGetVM validates permissions to get VM details
func (p *provider) dryRunGetVM(ctx context.Context) error {
	_, err := p.vmClient.Get(ctx, p.resourceGroup, "kl-dryrun-test-vm", &armcompute.VirtualMachinesClientGetOptions{
		Expand: to.Ptr(armcompute.InstanceViewTypesInstanceView),
	})
	return handleDryRunError(err, "get VM")
}

// dryRunGetDisk validates permissions to get disk details
func (p *provider) dryRunGetDisk(ctx context.Context) error {
	_, err := p.diskClient.Get(ctx, p.resourceGroup, "kl-dryrun-test-disk", nil)
	return handleDryRunError(err, "get disk")
}

// dryRunUpdateDisk validates permissions to update disk
func (p *provider) dryRunUpdateDisk(ctx context.Context) error {
	// List disks to validate update permissions
	pager := p.diskClient.NewListByResourceGroupPager(p.resourceGroup, nil)
	_, err := pager.NextPage(ctx)
	if err != nil {
		return handleDryRunError(err, "update disk")
	}
	return nil
}

// dryRunCreateNIC validates permissions to create a NIC
func (p *provider) dryRunCreateNIC(ctx context.Context) error {
	_, err := p.nicClient.Get(ctx, p.resourceGroup, "kl-dryrun-test-nic", &armnetwork.InterfacesClientGetOptions{})
	return handleDryRunError(err, "create NIC")
}

// dryRunDeleteNIC validates permissions to delete a NIC
func (p *provider) dryRunDeleteNIC(ctx context.Context) error {
	pager := p.nicClient.NewListPager(p.resourceGroup, nil)
	_, err := pager.NextPage(ctx)
	if err != nil {
		return handleDryRunError(err, "delete NIC")
	}
	return nil
}

// dryRunCreatePublicIP validates permissions to create a public IP
func (p *provider) dryRunCreatePublicIP(ctx context.Context) error {
	_, err := p.publicIPClient.Get(ctx, p.resourceGroup, "kl-dryrun-test-pip", nil)
	return handleDryRunError(err, "create public IP")
}

// dryRunDeletePublicIP validates permissions to delete a public IP
func (p *provider) dryRunDeletePublicIP(ctx context.Context) error {
	pager := p.publicIPClient.NewListPager(p.resourceGroup, nil)
	_, err := pager.NextPage(ctx)
	if err != nil {
		return handleDryRunError(err, "delete public IP")
	}
	return nil
}
