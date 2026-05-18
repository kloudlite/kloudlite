package workmachine

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/codingconcepts/env"
	"github.com/kloudlite/kloudlite/controllers/workmachine/cloud"
	"github.com/kloudlite/kloudlite/controllers/workmachine/cloud/aws"
	"github.com/kloudlite/kloudlite/controllers/workmachine/cloud/azure"
	"github.com/kloudlite/kloudlite/controllers/workmachine/cloud/gcp"
	ocicloud "github.com/kloudlite/kloudlite/controllers/workmachine/cloud/oci"
	"github.com/kloudlite/kloudlite/pkg/errors"
	v1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
)

func skipCloudPermissionValidation() bool {
	return strings.EqualFold(os.Getenv("KLOUDLITE_SKIP_CLOUD_PERMISSION_VALIDATION"), "true")
}

type awsProviderEnv struct {
	AWS_VPC_ID            string `env:"AWS_VPC_ID" required:"true"`
	AWS_SECURITY_GROUP_ID string `env:"AWS_SECURITY_GROUP_ID" required:"true"`
	AWS_REGION            string `env:"AWS_REGION" required:"true"`
}

type azureProviderEnv struct {
	AZURE_SUBSCRIPTION_ID string `env:"AZURE_SUBSCRIPTION_ID" required:"true"`
	AZURE_RESOURCE_GROUP  string `env:"AZURE_RESOURCE_GROUP" required:"true"`
	AZURE_LOCATION        string `env:"AZURE_LOCATION" required:"true"`
	AZURE_SUBNET_ID       string `env:"AZURE_SUBNET_ID" required:"true"`
	AZURE_NSG_ID          string `env:"AZURE_NSG_ID"`
}

type gcpProviderEnv struct {
	GCP_PROJECT    string `env:"GCP_PROJECT" required:"true"`
	GCP_REGION     string `env:"GCP_REGION" required:"true"`
	GCP_ZONE       string `env:"GCP_ZONE" required:"true"`
	GCP_NETWORK    string `env:"GCP_NETWORK" required:"true"`
	GCP_SUBNETWORK string `env:"GCP_SUBNETWORK" required:"true"`
}

type ociProviderEnv struct {
	OCI_COMPARTMENT string `env:"OCI_COMPARTMENT" required:"true"`
	OCI_REGION      string `env:"OCI_REGION" required:"true"`
	OCI_SUBNET_ID   string `env:"OCI_SUBNET_ID" required:"true"`
	OCI_NSG_ID      string `env:"OCI_NSG_ID" required:"true"`
}

func setupCloudProvider(ctx context.Context, cfg Env) (cloud.Provider, error) {
	switch cfg.CloudProvider {
	case v1.AWS:
		return setupAWSProvider(ctx, cfg)
	case v1.Azure:
		return setupAzureProvider(ctx, cfg)
	case v1.GCP:
		return setupGCPProvider(ctx, cfg)
	case v1.OCI:
		return setupOCIProvider(ctx, cfg)
	default:
		return nil, errors.New(fmt.Sprintf("unsupported cloud provider (%s)", cfg.CloudProvider))
	}
}

func setupAWSProvider(parent context.Context, cfg Env) (cloud.Provider, error) {
	var awsEnv awsProviderEnv
	if err := env.Set(&awsEnv); err != nil {
		return nil, errors.Wrap("failed to load env vars", err)
	}
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	p, err := aws.NewProvider(ctx, aws.ProviderArgs{
		Region:          awsEnv.AWS_REGION,
		VPC:             awsEnv.AWS_VPC_ID,
		SecurityGroupID: awsEnv.AWS_SECURITY_GROUP_ID,
		ResourceTags:    []aws.Tag{{Key: "kloudlite.io/installation-id", Value: cfg.KloudliteInstallationID}},
		K3sVersion:      cfg.K3sVersion,
		K3sURL:          cfg.K3sServerURL,
		K3sToken:        cfg.K3sAgentToken,
		HostedSubdomain: cfg.HostedSubdomain,
	})
	if err != nil {
		return nil, errors.Wrap("failed to create aws provider client", err)
	}
	return p, validateProviderPermissions(ctx, p)
}

func setupAzureProvider(parent context.Context, cfg Env) (cloud.Provider, error) {
	var azureEnv azureProviderEnv
	if err := env.Set(&azureEnv); err != nil {
		return nil, errors.Wrap("failed to load Azure env vars", err)
	}
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	p, err := azure.NewProvider(ctx, azure.ProviderArgs{
		SubscriptionID:         azureEnv.AZURE_SUBSCRIPTION_ID,
		ResourceGroup:          azureEnv.AZURE_RESOURCE_GROUP,
		Location:               azureEnv.AZURE_LOCATION,
		SubnetID:               azureEnv.AZURE_SUBNET_ID,
		NetworkSecurityGroupID: azureEnv.AZURE_NSG_ID,
		ResourceTags:           []azure.Tag{{Key: "kloudlite-installation-id", Value: cfg.KloudliteInstallationID}},
		K3sVersion:             cfg.K3sVersion,
		K3sURL:                 cfg.K3sServerURL,
		K3sToken:               cfg.K3sAgentToken,
		HostedSubdomain:        cfg.HostedSubdomain,
	})
	if err != nil {
		return nil, errors.Wrap("failed to create Azure provider client", err)
	}
	return p, validateProviderPermissions(ctx, p)
}

func setupGCPProvider(parent context.Context, cfg Env) (cloud.Provider, error) {
	var gcpEnv gcpProviderEnv
	if err := env.Set(&gcpEnv); err != nil {
		return nil, errors.Wrap("failed to load GCP env vars", err)
	}
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	p, err := gcp.NewProvider(ctx, gcp.ProviderArgs{
		Project:         gcpEnv.GCP_PROJECT,
		Region:          gcpEnv.GCP_REGION,
		Zone:            gcpEnv.GCP_ZONE,
		Network:         gcpEnv.GCP_NETWORK,
		Subnetwork:      gcpEnv.GCP_SUBNETWORK,
		ResourceTags:    []gcp.Tag{{Key: "kloudlite-installation-id", Value: cfg.KloudliteInstallationID}},
		K3sVersion:      cfg.K3sVersion,
		K3sURL:          cfg.K3sServerURL,
		K3sToken:        cfg.K3sAgentToken,
		HostedSubdomain: cfg.HostedSubdomain,
	})
	if err != nil {
		return nil, errors.Wrap("failed to create GCP provider client", err)
	}
	return p, validateProviderPermissions(ctx, p)
}

func setupOCIProvider(parent context.Context, cfg Env) (cloud.Provider, error) {
	var ociEnv ociProviderEnv
	if err := env.Set(&ociEnv); err != nil {
		return nil, errors.Wrap("failed to load OCI env vars", err)
	}
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	p, err := ocicloud.NewProvider(ctx, ocicloud.ProviderArgs{
		CompartmentID:   ociEnv.OCI_COMPARTMENT,
		Region:          ociEnv.OCI_REGION,
		SubnetID:        ociEnv.OCI_SUBNET_ID,
		NSGID:           ociEnv.OCI_NSG_ID,
		ResourceTags:    []ocicloud.Tag{{Key: "installation-id", Value: cfg.KloudliteInstallationID}},
		K3sVersion:      cfg.K3sVersion,
		K3sURL:          cfg.K3sServerURL,
		K3sToken:        cfg.K3sAgentToken,
		HostedSubdomain: cfg.HostedSubdomain,
	})
	if err != nil {
		return nil, errors.Wrap("failed to create OCI provider client", err)
	}
	return p, validateProviderPermissions(ctx, p)
}

func validateProviderPermissions(ctx context.Context, provider cloud.Provider) error {
	if skipCloudPermissionValidation() {
		return nil
	}
	return provider.ValidatePermissions(ctx)
}
