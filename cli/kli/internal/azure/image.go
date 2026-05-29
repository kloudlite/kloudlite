package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
)

// UbuntuImageReference contains the reference to Ubuntu 24.04 LTS image
type UbuntuImageReference struct {
	Publisher string
	Offer     string
	SKU       string
	Version   string
}

// GetUbuntu2404ImageReference returns the image reference for Ubuntu 24.04 LTS
func GetUbuntu2404ImageReference() *UbuntuImageReference {
	return &UbuntuImageReference{
		Publisher: "Canonical",
		Offer:     "ubuntu-24_04-lts",
		SKU:       "server",
		Version:   "latest",
	}
}

// FindUbuntuImage finds the latest Ubuntu 24.04 LTS image in the region
func FindUbuntuImage(ctx context.Context, cfg *AzureConfig) (*UbuntuImageReference, error) {
	client, err := armcompute.NewVirtualMachineImagesClient(cfg.SubscriptionID, cfg.Credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM images client: %w", err)
	}

	ref := GetUbuntu2404ImageReference()

	// Verify the image exists
	_, err = client.Get(ctx, cfg.Location, ref.Publisher, ref.Offer, ref.SKU, ref.Version, nil)
	if err != nil {
		// Try alternative SKU names
		alternativeSKUs := []string{"server-gen2", "server", "24_04-lts", "24_04-lts-gen2"}
		for _, sku := range alternativeSKUs {
			ref.SKU = sku
			_, err = client.Get(ctx, cfg.Location, ref.Publisher, ref.Offer, ref.SKU, ref.Version, nil)
			if err == nil {
				return ref, nil
			}
		}

		// Try listing available SKUs
		skus, listErr := client.ListSKUs(ctx, cfg.Location, ref.Publisher, ref.Offer, nil)
		if listErr == nil && len(skus.VirtualMachineImageResourceArray) > 0 {
			// Use the first available SKU
			ref.SKU = *skus.VirtualMachineImageResourceArray[0].Name
			return ref, nil
		}

		return nil, fmt.Errorf("Ubuntu 24.04 LTS image not found in region %s: %w", cfg.Location, err)
	}

	return ref, nil
}

// ToImageReference converts UbuntuImageReference to ARM ImageReference
func (r *UbuntuImageReference) ToImageReference() *armcompute.ImageReference {
	return &armcompute.ImageReference{
		Publisher: &r.Publisher,
		Offer:     &r.Offer,
		SKU:       &r.SKU,
		Version:   &r.Version,
	}
}

// String returns a human-readable string representation of the image reference
func (r *UbuntuImageReference) String() string {
	return fmt.Sprintf("%s/%s/%s/%s", r.Publisher, r.Offer, r.SKU, r.Version)
}
