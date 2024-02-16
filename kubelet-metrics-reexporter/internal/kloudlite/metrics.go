package kloudlite

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	fn "github.com/kloudlite/kubelet-metrics-reexporter/pkg/functions"
	"github.com/kloudlite/kubelet-metrics-reexporter/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubelet_stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

func writeMetric[T ~float64 | ~uint64](writer io.Writer, name MetricName, labels map[Label]string, value *T, timestamp metav1.Time) {
	if value == nil {
		return
	}
	keys := make([]Label, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}

	slices.SortFunc(keys, func(a, b Label) int {
		if LabelSortingOrder[a] < LabelSortingOrder[b] {
			return -1
		}
		return 1
	})

	s := make([]string, 0, len(labels))
	for _, key := range keys {
		if v := labels[key]; v != "" {
			s = append(s, fmt.Sprintf("%s=\"%s\"", key, labels[key]))
		}
	}

	fmt.Fprintf(writer, "%s{%s} %v %d\n", name, strings.Join(s, ","), *value, timestamp.UnixMilli())
}

type MetricsAggregator struct {
	extraTags map[string]string
	summary   *kubelet_stats.Summary
	node      *corev1.Node
	podsMap   k8s.PodsMap
}

func NewMetricsAggregator(ctx context.Context, kcli *k8s.Client, nodename string, extraTags map[string]string) (*MetricsAggregator, error) {
	summary, err := kcli.StatsSummary(ctx, nodename)
	if err != nil {
		return nil, err
	}

	node, err := kcli.GetNode(ctx, nodename)
	if err != nil {
		return nil, err
	}

	pods, err := kcli.ListPodsOnNode(ctx, nodename)
	if err != nil {
		return nil, err
	}

	pm := k8s.ToPodsMap(pods)

	return &MetricsAggregator{
		extraTags: extraTags,
		summary:   summary,
		node:      node,
		podsMap:   pm,
	}, nil
}

func (m *MetricsAggregator) WriteNodeMetrics(writer io.Writer) error {
	nodeLabels := map[Label]string{
		KloudliteAccountName: "--unknown--",
		KloudliteClusterName: "--unknown--",
		KloudliteTrackingId:  "--unknown--",

		NodeName: m.summary.Node.NodeName,
	}

	for k, v := range m.extraTags {
		nodeLabels[Label(k)] = v
	}

	if cpu := m.summary.Node.CPU; cpu != nil {
		cpuUtilization := (float64(*m.summary.Node.CPU.UsageNanoCores) / (1e9 * m.node.Status.Allocatable.Cpu().AsApproximateFloat64())) * 100

		writeMetric(writer, NodeCpuUtilization, nodeLabels, &cpuUtilization, cpu.Time)
		writeMetric(writer, NodeCpuUsage, nodeLabels, fn.New(float64(*m.summary.Node.CPU.UsageNanoCores)/1e6), cpu.Time)
	}

	if mem := m.summary.Node.Memory; mem != nil {
		writeMetric(writer, NodeMemoryUsed, nodeLabels, mem.WorkingSetBytes, mem.Time)
		writeMetric(writer, NodeMemoryAvailable, nodeLabels, mem.AvailableBytes, mem.Time)
	}

	if fs := m.summary.Node.Fs; fs != nil {
		writeMetric(writer, NodeStorageUsed, nodeLabels, fs.UsedBytes, fs.Time)
		writeMetric(writer, NodeStorageAvail, nodeLabels, fs.AvailableBytes, fs.Time)
		writeMetric(writer, NodeStorageCapacity, nodeLabels, fs.CapacityBytes, fs.Time)
	}

	if network := m.summary.Node.Network; network != nil {
		for _, niface := range network.Interfaces {
			labels := fn.MapMerge(nodeLabels, map[Label]string{NetworkInterface: niface.Name})
			writeMetric(writer, NodeNetworkRead, labels, niface.RxBytes, network.Time)
			writeMetric(writer, NodeNetworkReadErrors, labels, niface.RxErrors, network.Time)
			writeMetric(writer, NodeNetworkWrite, labels, niface.TxBytes, network.Time)
			writeMetric(writer, NodeNetworkWriteErrors, labels, niface.TxErrors, network.Time)
		}
	}

	return nil
}

func (m *MetricsAggregator) WritePodMetrics(writer io.Writer) error {
	for i := range m.summary.Pods {
		podname := m.summary.Pods[i].PodRef.Name
		podns := m.summary.Pods[i].PodRef.Namespace

		commonLabels := map[Label]string{
			KloudliteAccountName: m.podsMap.PodAccountName(podns, podname),
			KloudliteClusterName: m.podsMap.PodClusterName(podns, podname),
			KloudliteTrackingId:  m.podsMap.PodTrackingId(podns, podname),

			PodName:      podname,
			PodNamespace: podns,
		}

		for k, v := range m.extraTags {
			commonLabels[Label(k)] = v
		}

		if cpu := m.summary.Pods[i].CPU; cpu != nil {
			if cpu.UsageNanoCores != nil {
				writeMetric(writer, PodCpuUsage, commonLabels, fn.New(float64(*cpu.UsageNanoCores)/1e6), cpu.Time)
			}
		}

		if mem := m.summary.Pods[i].Memory; mem != nil {
			writeMetric(writer, PodMemoryUsed, commonLabels, mem.WorkingSetBytes, mem.Time)
			writeMetric(writer, PodMemoryAvailable, commonLabels, mem.AvailableBytes, mem.Time)
		}

		for _, vs := range m.summary.Pods[i].VolumeStats {
			if pvc := vs.PVCRef; pvc != nil {
				labels := fn.MapMerge(commonLabels, map[Label]string{
					PVCName:      pvc.Name,
					PVCNamespace: pvc.Namespace,
				})
				writeMetric(writer, PodPvcStorageUsed, labels, vs.UsedBytes, vs.Time)
				writeMetric(writer, PodPvcStorageAvail, labels, vs.AvailableBytes, vs.Time)
				writeMetric(writer, PodPvcStorageCapacity, labels, vs.CapacityBytes, vs.Time)
			}
		}

		if network := m.summary.Pods[i].Network; network != nil {
			for _, niface := range network.Interfaces {
				labels := fn.MapMerge(commonLabels, map[Label]string{
					NetworkInterface: niface.Name,
				})

				writeMetric(writer, PodNetworkRead, labels, niface.RxBytes, network.Time)
				writeMetric(writer, PodNetworkReadErrors, labels, niface.RxErrors, network.Time)
				writeMetric(writer, PodNetworkWrite, labels, niface.TxBytes, network.Time)
				writeMetric(writer, PodNetworkWriteErrors, labels, niface.TxErrors, network.Time)
			}
		}
	}

	return nil
}
