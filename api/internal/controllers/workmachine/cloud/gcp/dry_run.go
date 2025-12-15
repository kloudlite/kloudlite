package gcp

import (
	"context"
	"strings"

	"cloud.google.com/go/compute/apiv1/computepb"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
)

// GCP Permission Validation Approach
//
// Unlike AWS, GCP doesn't have a native "dry-run" mode for API calls.
// Instead, we validate permissions by attempting read operations that
// require the same IAM permissions as write operations.
//
// Permission validation strategy:
// 1. Try list/get operations on resources (requires compute.instances.list, etc.)
// 2. Check error responses for permission issues
// 3. Permission errors (403/PERMISSION_DENIED) = fail
// 4. Success or resource-not-found = pass (permission check succeeded)
//
// Required GCP IAM Roles for WorkMachine operations:
// - roles/compute.instanceAdmin.v1 (create, delete, start, stop, reset instances)
// - roles/compute.storageAdmin (attach/resize disks)

// handlePermissionError checks if an error indicates a permission problem
// Returns nil if permissions are OK (including "not found" errors which indicate permission was granted)
// Returns the error if it's a permission-related error
func handlePermissionError(err error, action string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Permission errors indicate IAM is missing
	if strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "PERMISSION_DENIED") ||
		strings.Contains(errStr, "accessNotConfigured") ||
		strings.Contains(errStr, "forbidden") {
		return err
	}

	// Not found errors mean permission check passed (resource just doesn't exist)
	if strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "notFound") ||
		strings.Contains(errStr, "NOT_FOUND") {
		return nil
	}

	// For other errors, check if they're permission-related
	if strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "Unauthorized") {
		return err
	}

	// Non-permission errors are OK for validation purposes
	return nil
}

// checkListInstances validates compute.instances.list permission
func (p *provider) checkListInstances(ctx context.Context) error {
	req := &computepb.ListInstancesRequest{
		Project:    p.Project,
		Zone:       p.Zone,
		MaxResults: fn.Ptr(uint32(1)),
	}

	it := p.instancesClient.List(ctx, req)
	_, err := it.Next()

	// iterator.Done is not an error, just means empty list
	if err != nil && !strings.Contains(err.Error(), "iterator done") {
		return handlePermissionError(err, "compute.instances.list")
	}

	return nil
}

// checkListDisks validates compute.disks.list permission
func (p *provider) checkListDisks(ctx context.Context) error {
	req := &computepb.ListDisksRequest{
		Project:    p.Project,
		Zone:       p.Zone,
		MaxResults: fn.Ptr(uint32(1)),
	}

	it := p.disksClient.List(ctx, req)
	_, err := it.Next()

	// iterator.Done is not an error, just means empty list
	if err != nil && !strings.Contains(err.Error(), "iterator done") {
		return handlePermissionError(err, "compute.disks.list")
	}

	return nil
}
