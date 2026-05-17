package gcp

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/serviceusage/v1"
)

// Required APIs for Kloudlite installation
var RequiredAPIs = []string{
	"compute.googleapis.com",
	"iam.googleapis.com",
	"storage.googleapis.com",
	"iamcredentials.googleapis.com",
	"cloudresourcemanager.googleapis.com",
}

// EnableRequiredAPIs enables all required GCP APIs for the project
func EnableRequiredAPIs(ctx context.Context, project string) error {
	service, err := serviceusage.NewService(ctx)
	if err != nil {
		return fmt.Errorf("failed to create service usage client: %w", err)
	}

	for _, api := range RequiredAPIs {
		serviceName := fmt.Sprintf("projects/%s/services/%s", project, api)

		// Check if already enabled
		svc, err := service.Services.Get(serviceName).Context(ctx).Do()
		if err == nil && svc.State == "ENABLED" {
			continue
		}

		// Enable the API
		op, err := service.Services.Enable(serviceName, &serviceusage.EnableServiceRequest{}).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("failed to enable %s: %w", api, err)
		}

		// Wait for operation to complete
		for !op.Done {
			time.Sleep(2 * time.Second)
			op, err = service.Operations.Get(op.Name).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("failed to check operation status for %s: %w", api, err)
			}
		}

		if op.Error != nil {
			return fmt.Errorf("failed to enable %s: %s", api, op.Error.Message)
		}
	}

	return nil
}
