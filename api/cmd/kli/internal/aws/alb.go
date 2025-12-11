package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elbTypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

// ALBInfo contains information about the created ALB
type ALBInfo struct {
	ARN     string
	DNSName string
}

// CreateALB creates an Application Load Balancer
func CreateALB(ctx context.Context, cfg aws.Config, installationKey, vpcID string, subnetIDs []string, sgID string) (*ALBInfo, error) {
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)
	albName := fmt.Sprintf("kl-%s-alb", installationKey)

	// Ensure we have at least 2 subnets from different AZs
	if len(subnetIDs) < 2 {
		return nil, fmt.Errorf("ALB requires at least 2 subnets from different availability zones, got %d", len(subnetIDs))
	}

	result, err := elbClient.CreateLoadBalancer(ctx, &elasticloadbalancingv2.CreateLoadBalancerInput{
		Name:           aws.String(albName),
		Subnets:        subnetIDs,
		SecurityGroups: []string{sgID},
		Scheme:         elbTypes.LoadBalancerSchemeEnumInternetFacing,
		Type:           elbTypes.LoadBalancerTypeEnumApplication,
		IpAddressType:  elbTypes.IpAddressTypeIpv4,
		Tags: []elbTypes.Tag{
			{Key: aws.String("Name"), Value: aws.String(albName)},
			{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
			{Key: aws.String("Project"), Value: aws.String("kloudlite")},
			{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
			{Key: aws.String("kloudlite.io/installation-id"), Value: aws.String(installationKey)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ALB: %w", err)
	}

	if len(result.LoadBalancers) == 0 {
		return nil, fmt.Errorf("no load balancer returned after creation")
	}

	alb := result.LoadBalancers[0]
	return &ALBInfo{
		ARN:     *alb.LoadBalancerArn,
		DNSName: *alb.DNSName,
	}, nil
}

// CreateTargetGroup creates a target group for the ALB
func CreateTargetGroup(ctx context.Context, cfg aws.Config, installationKey, vpcID string) (string, error) {
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)
	tgName := fmt.Sprintf("kl-%s-tg", installationKey)

	result, err := elbClient.CreateTargetGroup(ctx, &elasticloadbalancingv2.CreateTargetGroupInput{
		Name:                       aws.String(tgName),
		Protocol:                   elbTypes.ProtocolEnumHttp,
		Port:                       aws.Int32(80),
		VpcId:                      aws.String(vpcID),
		TargetType:                 elbTypes.TargetTypeEnumInstance,
		HealthCheckEnabled:         aws.Bool(true),
		HealthCheckPath:            aws.String("/"),
		HealthCheckProtocol:        elbTypes.ProtocolEnumHttp,
		HealthCheckPort:            aws.String("traffic-port"),
		HealthyThresholdCount:      aws.Int32(2),
		UnhealthyThresholdCount:    aws.Int32(3),
		HealthCheckIntervalSeconds: aws.Int32(30),
		HealthCheckTimeoutSeconds:  aws.Int32(5),
		Tags: []elbTypes.Tag{
			{Key: aws.String("Name"), Value: aws.String(tgName)},
			{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
			{Key: aws.String("Project"), Value: aws.String("kloudlite")},
			{Key: aws.String("Purpose"), Value: aws.String("kloudlite-installation")},
			{Key: aws.String("kloudlite.io/installation-id"), Value: aws.String(installationKey)},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create target group: %w", err)
	}

	if len(result.TargetGroups) == 0 {
		return "", fmt.Errorf("no target group returned after creation")
	}

	return *result.TargetGroups[0].TargetGroupArn, nil
}

// RegisterTargets registers EC2 instances with the target group
func RegisterTargets(ctx context.Context, cfg aws.Config, tgARN string, instanceIDs ...string) error {
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	var targets []elbTypes.TargetDescription
	for _, id := range instanceIDs {
		targets = append(targets, elbTypes.TargetDescription{
			Id:   aws.String(id),
			Port: aws.Int32(80),
		})
	}

	_, err := elbClient.RegisterTargets(ctx, &elasticloadbalancingv2.RegisterTargetsInput{
		TargetGroupArn: aws.String(tgARN),
		Targets:        targets,
	})
	if err != nil {
		return fmt.Errorf("failed to register targets: %w", err)
	}

	return nil
}

// CreateHTTPSListener creates an HTTPS listener with TLS termination
func CreateHTTPSListener(ctx context.Context, cfg aws.Config, albARN, tgARN, certARN string) (string, error) {
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	result, err := elbClient.CreateListener(ctx, &elasticloadbalancingv2.CreateListenerInput{
		LoadBalancerArn: aws.String(albARN),
		Protocol:        elbTypes.ProtocolEnumHttps,
		Port:            aws.Int32(443),
		SslPolicy:       aws.String("ELBSecurityPolicy-TLS13-1-2-2021-06"),
		Certificates: []elbTypes.Certificate{
			{CertificateArn: aws.String(certARN)},
		},
		DefaultActions: []elbTypes.Action{
			{
				Type:           elbTypes.ActionTypeEnumForward,
				TargetGroupArn: aws.String(tgARN),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create HTTPS listener: %w", err)
	}

	if len(result.Listeners) == 0 {
		return "", fmt.Errorf("no listener returned after creation")
	}

	return *result.Listeners[0].ListenerArn, nil
}

// CreateHTTPRedirectListener creates an HTTP listener that redirects to HTTPS
func CreateHTTPRedirectListener(ctx context.Context, cfg aws.Config, albARN string) (string, error) {
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	result, err := elbClient.CreateListener(ctx, &elasticloadbalancingv2.CreateListenerInput{
		LoadBalancerArn: aws.String(albARN),
		Protocol:        elbTypes.ProtocolEnumHttp,
		Port:            aws.Int32(80),
		DefaultActions: []elbTypes.Action{
			{
				Type: elbTypes.ActionTypeEnumRedirect,
				RedirectConfig: &elbTypes.RedirectActionConfig{
					Protocol:   aws.String("HTTPS"),
					Port:       aws.String("443"),
					StatusCode: elbTypes.RedirectActionStatusCodeEnumHttp301,
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP redirect listener: %w", err)
	}

	if len(result.Listeners) == 0 {
		return "", fmt.Errorf("no listener returned after creation")
	}

	return *result.Listeners[0].ListenerArn, nil
}

// WaitForALBActive waits for the ALB to become active
func WaitForALBActive(ctx context.Context, cfg aws.Config, albARN string) error {
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	timeout := 5 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		result, err := elbClient.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{
			LoadBalancerArns: []string{albARN},
		})
		if err != nil {
			return fmt.Errorf("failed to describe ALB: %w", err)
		}

		if len(result.LoadBalancers) > 0 {
			state := result.LoadBalancers[0].State
			if state != nil && state.Code == elbTypes.LoadBalancerStateEnumActive {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			// Continue polling
		}
	}

	return fmt.Errorf("ALB did not become active within %v", timeout)
}

// DeleteALB deletes an ALB by installation key
func DeleteALB(ctx context.Context, cfg aws.Config, installationKey string) error {
	albARN, err := FindALBByInstallationKey(ctx, cfg, installationKey)
	if err != nil {
		return err
	}
	if albARN == "" {
		return nil // ALB doesn't exist
	}

	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	// First delete all listeners
	listeners, err := elbClient.DescribeListeners(ctx, &elasticloadbalancingv2.DescribeListenersInput{
		LoadBalancerArn: aws.String(albARN),
	})
	if err == nil {
		for _, listener := range listeners.Listeners {
			_, _ = elbClient.DeleteListener(ctx, &elasticloadbalancingv2.DeleteListenerInput{
				ListenerArn: listener.ListenerArn,
			})
		}
	}

	// Delete the ALB
	_, err = elbClient.DeleteLoadBalancer(ctx, &elasticloadbalancingv2.DeleteLoadBalancerInput{
		LoadBalancerArn: aws.String(albARN),
	})
	if err != nil {
		return fmt.Errorf("failed to delete ALB: %w", err)
	}

	// Wait for ALB to be deleted
	return WaitForALBDeleted(ctx, cfg, albARN)
}

// WaitForALBDeleted waits for an ALB to be fully deleted
func WaitForALBDeleted(ctx context.Context, cfg aws.Config, albARN string) error {
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	timeout := 5 * time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		result, err := elbClient.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{
			LoadBalancerArns: []string{albARN},
		})
		if err != nil {
			// ALB not found means it's deleted
			return nil
		}

		if len(result.LoadBalancers) == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			// Continue polling
		}
	}

	return fmt.Errorf("ALB deletion timed out after %v", timeout)
}

// DeleteTargetGroup deletes a target group by installation key
func DeleteTargetGroup(ctx context.Context, cfg aws.Config, installationKey string) error {
	tgARN, err := FindTargetGroupByInstallationKey(ctx, cfg, installationKey)
	if err != nil {
		return err
	}
	if tgARN == "" {
		return nil // Target group doesn't exist
	}

	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	_, err = elbClient.DeleteTargetGroup(ctx, &elasticloadbalancingv2.DeleteTargetGroupInput{
		TargetGroupArn: aws.String(tgARN),
	})
	if err != nil {
		return fmt.Errorf("failed to delete target group: %w", err)
	}

	return nil
}

// FindALBByInstallationKey finds an ALB by installation key tag
func FindALBByInstallationKey(ctx context.Context, cfg aws.Config, installationKey string) (string, error) {
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	result, err := elbClient.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
	if err != nil {
		return "", fmt.Errorf("failed to describe load balancers: %w", err)
	}

	for _, alb := range result.LoadBalancers {
		// Get tags for this ALB
		tagsResult, err := elbClient.DescribeTags(ctx, &elasticloadbalancingv2.DescribeTagsInput{
			ResourceArns: []string{*alb.LoadBalancerArn},
		})
		if err != nil {
			continue
		}

		for _, tagDesc := range tagsResult.TagDescriptions {
			for _, tag := range tagDesc.Tags {
				if tag.Key != nil && *tag.Key == "kloudlite.io/installation-id" &&
					tag.Value != nil && *tag.Value == installationKey {
					return *alb.LoadBalancerArn, nil
				}
			}
		}
	}

	return "", nil // Not found
}

// FindTargetGroupByInstallationKey finds a target group by installation key tag
func FindTargetGroupByInstallationKey(ctx context.Context, cfg aws.Config, installationKey string) (string, error) {
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	result, err := elbClient.DescribeTargetGroups(ctx, &elasticloadbalancingv2.DescribeTargetGroupsInput{})
	if err != nil {
		return "", fmt.Errorf("failed to describe target groups: %w", err)
	}

	for _, tg := range result.TargetGroups {
		// Get tags for this target group
		tagsResult, err := elbClient.DescribeTags(ctx, &elasticloadbalancingv2.DescribeTagsInput{
			ResourceArns: []string{*tg.TargetGroupArn},
		})
		if err != nil {
			continue
		}

		for _, tagDesc := range tagsResult.TagDescriptions {
			for _, tag := range tagDesc.Tags {
				if tag.Key != nil && *tag.Key == "kloudlite.io/installation-id" &&
					tag.Value != nil && *tag.Value == installationKey {
					return *tg.TargetGroupArn, nil
				}
			}
		}
	}

	return "", nil // Not found
}

// GetALBDNSName gets the DNS name of an ALB by its ARN
func GetALBDNSName(ctx context.Context, cfg aws.Config, albARN string) (string, error) {
	elbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	result, err := elbClient.DescribeLoadBalancers(ctx, &elasticloadbalancingv2.DescribeLoadBalancersInput{
		LoadBalancerArns: []string{albARN},
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe ALB: %w", err)
	}

	if len(result.LoadBalancers) == 0 {
		return "", fmt.Errorf("ALB not found")
	}

	return *result.LoadBalancers[0].DNSName, nil
}
