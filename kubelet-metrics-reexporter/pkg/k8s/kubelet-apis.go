package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	kubelet_stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

// MetricsResource is API for kubelet API /stats/summary, returning k8s.io/kubelet/pkg/apis/stats/v1alpha1.Summary
func (c *Client) StatsSummary(ctx context.Context, nodename string) (*kubelet_stats.Summary, error) {
	req := c.RESTClient().Get().AbsPath(fmt.Sprintf("/api/v1/nodes/%s/proxy/stats/summary", nodename))

	b, err := req.DoRaw(ctx)
	if err != nil {
		return nil, err
	}

	var summary kubelet_stats.Summary
	if err := json.Unmarshal(b, &summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

// MetricsResource is API for kubelet API /stats/summary, returning json response
func (c *Client) StatsSummaryRaw(ctx context.Context, nodename string) ([]byte, error) {
	req := c.RESTClient().Get().AbsPath(fmt.Sprintf("/api/v1/nodes/%s/proxy/stats/summary", nodename))

	return req.DoRaw(ctx)
}

// MetricsResource is API for kubelet API /metrics/resource
func (c *Client) MetricsResource(ctx context.Context, nodename string) ([]byte, error) {
	req := c.RESTClient().Get().AbsPath(fmt.Sprintf("/api/v1/nodes/%s/proxy/metrics/resource", nodename))
	return req.DoRaw(ctx)
}

// Metrics is API for kubelet API /metrics
func (c *Client) Metrics(ctx context.Context, nodename string) ([]byte, error) {
	req := c.RESTClient().Get().AbsPath(fmt.Sprintf("/api/v1/nodes/%s/proxy/metrics", nodename))
	return req.DoRaw(ctx)
}
