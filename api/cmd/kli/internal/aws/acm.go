package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
)

// ValidationRecord represents a DNS validation record for ACM certificate
type ValidationRecord struct {
	Name  string
	Value string
	Type  string
}

// RequestCertificate requests a new ACM certificate for the given domain (idempotent - returns existing if found)
// Returns the certificate ARN
func RequestCertificate(ctx context.Context, cfg aws.Config, domain string, installationKey string) (string, error) {
	// Check if certificate already exists for this installation
	existingARN, err := FindCertificateByInstallationKey(ctx, cfg, installationKey)
	if err == nil && existingARN != "" {
		return existingARN, nil
	}

	acmClient := acm.NewFromConfig(cfg)

	// Request certificate for domain and wildcard
	result, err := acmClient.RequestCertificate(ctx, &acm.RequestCertificateInput{
		DomainName: aws.String(domain),
		SubjectAlternativeNames: []string{
			domain,
			fmt.Sprintf("*.%s", domain),
		},
		ValidationMethod: types.ValidationMethodDns,
		Tags: []types.Tag{
			{Key: aws.String("Name"), Value: aws.String(fmt.Sprintf("kl-%s-cert", installationKey))},
			{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
			{Key: aws.String("Project"), Value: aws.String("kloudlite")},
			{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
			{Key: aws.String("kloudlite.io/installation-id"), Value: aws.String(installationKey)},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to request certificate: %w", err)
	}

	return *result.CertificateArn, nil
}

// GetValidationRecords retrieves the DNS validation records for an ACM certificate
// Returns the records needed for DNS validation
func GetValidationRecords(ctx context.Context, cfg aws.Config, certARN string) ([]ValidationRecord, error) {
	acmClient := acm.NewFromConfig(cfg)

	// Wait for validation records to be available (may take a few seconds)
	var records []ValidationRecord
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		result, err := acmClient.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
			CertificateArn: aws.String(certARN),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe certificate: %w", err)
		}

		if len(result.Certificate.DomainValidationOptions) > 0 {
			for _, opt := range result.Certificate.DomainValidationOptions {
				if opt.ResourceRecord != nil {
					records = append(records, ValidationRecord{
						Name:  *opt.ResourceRecord.Name,
						Value: *opt.ResourceRecord.Value,
						Type:  string(opt.ResourceRecord.Type),
					})
				}
			}
			if len(records) > 0 {
				// Dedupe records (domain and wildcard often share the same validation record)
				return dedupeValidationRecords(records), nil
			}
		}

		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("validation records not available after %d seconds", maxAttempts*2)
}

// dedupeValidationRecords removes duplicate validation records
func dedupeValidationRecords(records []ValidationRecord) []ValidationRecord {
	seen := make(map[string]bool)
	var result []ValidationRecord
	for _, r := range records {
		key := r.Name + ":" + r.Value
		if !seen[key] {
			seen[key] = true
			result = append(result, r)
		}
	}
	return result
}

// WaitForValidation waits for the ACM certificate to be validated
func WaitForValidation(ctx context.Context, cfg aws.Config, certARN string, timeout time.Duration) error {
	acmClient := acm.NewFromConfig(cfg)

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		result, err := acmClient.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
			CertificateArn: aws.String(certARN),
		})
		if err != nil {
			return fmt.Errorf("failed to describe certificate: %w", err)
		}

		status := result.Certificate.Status
		switch status {
		case types.CertificateStatusIssued:
			return nil
		case types.CertificateStatusFailed, types.CertificateStatusRevoked:
			return fmt.Errorf("certificate validation failed with status: %s", status)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			// Continue polling
		}
	}

	return fmt.Errorf("certificate validation timed out after %v", timeout)
}

// GetCertificateStatus returns the current status of an ACM certificate
func GetCertificateStatus(ctx context.Context, cfg aws.Config, certARN string) (string, error) {
	acmClient := acm.NewFromConfig(cfg)

	result, err := acmClient.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
		CertificateArn: aws.String(certARN),
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe certificate: %w", err)
	}

	return string(result.Certificate.Status), nil
}

// DeleteCertificate deletes an ACM certificate with retry for in-use errors
func DeleteCertificate(ctx context.Context, cfg aws.Config, certARN string) error {
	acmClient := acm.NewFromConfig(cfg)

	// Retry with backoff for ResourceInUseException
	// This handles the case where the ALB listener is still being deleted
	maxRetries := 12
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			waitTime := time.Duration(10+min(i*5, 30)) * time.Second
			fmt.Printf("    [%s] Certificate still in use, retry %d/%d, waiting %ds...\n",
				time.Now().Format("15:04:05"), i, maxRetries-1, int(waitTime.Seconds()))
			time.Sleep(waitTime)
		}

		_, err := acmClient.DeleteCertificate(ctx, &acm.DeleteCertificateInput{
			CertificateArn: aws.String(certARN),
		})
		if err == nil {
			return nil
		}

		// Check if it's a ResourceInUseException (certificate still attached to a listener)
		errMsg := err.Error()
		if i < maxRetries-1 && (strings.Contains(errMsg, "ResourceInUseException") || strings.Contains(errMsg, "in use")) {
			continue
		}

		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	return fmt.Errorf("failed to delete certificate after %d retries: still in use", maxRetries)
}

// FindCertificateByInstallationKey finds an ACM certificate by installation key tag
func FindCertificateByInstallationKey(ctx context.Context, cfg aws.Config, installationKey string) (string, error) {
	acmClient := acm.NewFromConfig(cfg)

	result, err := acmClient.ListCertificates(ctx, &acm.ListCertificatesInput{})
	if err != nil {
		return "", fmt.Errorf("failed to list certificates: %w", err)
	}

	for _, cert := range result.CertificateSummaryList {
		// Get tags for this certificate
		tagsResult, err := acmClient.ListTagsForCertificate(ctx, &acm.ListTagsForCertificateInput{
			CertificateArn: cert.CertificateArn,
		})
		if err != nil {
			continue
		}

		for _, tag := range tagsResult.Tags {
			if tag.Key != nil && *tag.Key == "kloudlite.io/installation-id" &&
				tag.Value != nil && *tag.Value == installationKey {
				return *cert.CertificateArn, nil
			}
		}
	}

	return "", nil // Not found
}
