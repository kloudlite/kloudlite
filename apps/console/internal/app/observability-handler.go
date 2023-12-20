package app

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
)

type ObservabilityArgs struct {
	AccountName string `json:"account_name"`
	ClusterName string `json:"cluster_name"`

	ResourceName      string `json:"resource_name"`
	ResourceNamespace string `json:"resource_namespace"`
	ResourceType      string `json:"resource_type"`
	WorkspaceName     string `json:"workspace_name"`
	ProjectName       string `json:"project_name"`

	JobName      string `json:"job_name"`
	JobNamespace string `json:"job_namespace"`

	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
}

func (args *ObservabilityArgs) Validate() (bool, error) {
	errorMsgs := make([]string, 0, 5)
	if args.AccountName == "" {
		errorMsgs = append(errorMsgs, "account_name is required")
	}

	if args.ClusterName == "" {
		errorMsgs = append(errorMsgs, "cluster_name is required")
	}

	hasResource := true
	if (args.ResourceName != "" && args.ResourceNamespace == "") || (args.ResourceName == "" && args.ResourceNamespace != "") {
		hasResource = false
		errorMsgs = append(errorMsgs, "resource_name and resource_namespace must be provided in pair")
	}

	if !hasResource && (args.WorkspaceName == "" && args.ProjectName == "") {
		errorMsgs = append(errorMsgs, "workspace_name/project_name is required")
	}

	if len(errorMsgs) > 0 {
		b, err := json.Marshal(map[string]any{"error": errorMsgs})
		if err != nil {
			return false, errors.NewE(err)
		}
		return false, errors.Newf(string(b))
	}
	return true, nil
}

type PromMetricsType string

const (
	Cpu                PromMetricsType = "cpu"
	Memory             PromMetricsType = "memory"
	DiskIO             PromMetricsType = "disk-io"
	NetworkReceived    PromMetricsType = "network-received"
	NetworkTransmitted PromMetricsType = "network-transmitted"
)

type ObservabilityLabel string

const (
	AccountName ObservabilityLabel = "kl_account_name"
	ClusterName ObservabilityLabel = "kl_cluster_name"

	ResourceName      ObservabilityLabel = "kl_resource_name"
	ResourceType      ObservabilityLabel = "kl_resource_type"
	ResourceNamespace ObservabilityLabel = "kl_resource_namespace"
	ResourceComponent ObservabilityLabel = "kl_resource_component"

	ProjectName            ObservabilityLabel = "kl_project_name"
	ProjectTargetNamespace ObservabilityLabel = "kl_project_target_ns"

	WorkspaceName     ObservabilityLabel = "kl_workspace_name"
	WorkspaceTargetNs ObservabilityLabel = "kl_workspace_target_ns"
)

func buildPromQuery(resType PromMetricsType, filters map[ObservabilityLabel]string) (string, error) {
	// switch resType {
	// case Memory:
	// 	return fmt.Sprintf(`sum(avg_over_time(container_memory_working_set_bytes{namespace="%s",pod=~"%s.*",container!="POD",image!=""}[30s]))/1024/1024`, namespace, name), nil
	// case Cpu:
	// 	return fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s", pod=~"%s.*", image!="", container!="POD"}[1m])) * 1000`, namespace, name), nil
	// case NetworkTransmitted:
	// 	return fmt.Sprintf(""), nil
	// }

	tags := make([]string, 0, len(filters))

	for k, v := range filters {
		if v != "" {
			tags = append(tags, fmt.Sprintf(`%s=%q`, k, v))
		}
	}

	switch resType {
	case Memory:
		return fmt.Sprintf(`sum(avg_over_time(pod_memory_working_set_bytes{%s}[1m]))/1024/1024`, strings.Join(tags, ",")), nil
	case Cpu:
		return fmt.Sprintf(`sum(rate(pod_cpu_usage_seconds_total{%s}[2m])) * 1000`, strings.Join(tags, ",")), nil
	}

	return "", nil
}

const (
	DefaultStepSize = 700 // 15 minutes
)

func queryProm(promAddr string, resType PromMetricsType, filters map[ObservabilityLabel]string, startTime *time.Time, endTime *time.Time, writer io.Writer) error {
	promQuery, err := buildPromQuery(resType, filters)
	if err != nil {
		return errors.NewE(err)
	}

	pu, err := url.Parse(promAddr)
	if err != nil {
		return errors.NewEf(err, "failed to parser promAddr into *url.URL")
	}

	u := pu.JoinPath("/api/v1/query_range")

	qp := u.Query()
	qp.Add("query", promQuery)

	t := time.Now()
	if startTime == nil {
		startTime = fn.New(t.Add(-2 * 24 * time.Hour))
	}
	if endTime == nil {
		endTime = &t
	}

	qp.Add("start", fmt.Sprintf("%d", startTime.Unix()))
	qp.Add("end", fmt.Sprintf("%d", endTime.Unix()))
	// qp.Add("step", "700") // 15 minute
	qp.Add("step", "345") // 15 minute

	u.RawQuery = qp.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return errors.NewE(err)
	}

	fmt.Printf("[DEBUG]: prometheus actual request: %s\n", req.URL.String())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.NewE(err)
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Newf("incorrect status code, expected %d, got %d", http.StatusOK, resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.NewE(err)
	}

	_, err = writer.Write(b)
	return errors.NewE(err)
}
