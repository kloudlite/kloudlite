package kloudlite

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/nxtcoder17/kubelet-metrics-reexporter/internal/types"
	fn "github.com/nxtcoder17/kubelet-metrics-reexporter/pkg/functions"
	corev1 "k8s.io/api/core/v1"
	kubelet_stats "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
)

func writeMetric[T ~float64 | ~uint64](writer io.Writer, name MetricName, labels map[Label]string, value *T, timestamp time.Time) {
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

func Metrics(summary kubelet_stats.Summary, node *corev1.Node, podsMap types.PodsMap, writer io.Writer) error {
	nodeLabels := map[Label]string{
		KloudliteAccountName: "",
		KloudliteClusterName: "",
		KloudliteTrackingId:  "",

		NodeName: summary.Node.NodeName,
	}

	if cpu := summary.Node.CPU; cpu != nil {
		cpuUtilization := (float64(*summary.Node.CPU.UsageNanoCores) / (1e9 * node.Status.Allocatable.Cpu().AsApproximateFloat64())) * 100

		writeMetric(writer, NodeCpuUtilization, nodeLabels, &cpuUtilization, cpu.Time.Time)
		writeMetric(writer, NodeCpuUsage, nodeLabels, fn.New(float64(*summary.Node.CPU.UsageNanoCores)/1e6), cpu.Time.Time)
	}

	if mem := summary.Node.Memory; mem != nil {
		writeMetric(writer, NodeMemoryUsed, nodeLabels, mem.WorkingSetBytes, mem.Time.Time)
		writeMetric(writer, NodeMemoryAvailable, nodeLabels, mem.AvailableBytes, mem.Time.Time)
	}

	if fs := summary.Node.Fs; fs != nil {
		writeMetric(writer, NodeStorageUsed, nodeLabels, fs.UsedBytes, fs.Time.Time)
		writeMetric(writer, NodeStorageAvail, nodeLabels, fs.AvailableBytes, fs.Time.Time)
		writeMetric(writer, NodeStorageCapacity, nodeLabels, fs.CapacityBytes, fs.Time.Time)
	}

	if network := summary.Node.Network; network != nil {
		for _, niface := range network.Interfaces {
			labels := fn.MapMerge(nodeLabels, map[Label]string{NetworkInterface: niface.Name})
			writeMetric(writer, NodeNetworkRead, labels, niface.RxBytes, network.Time.Time)
			writeMetric(writer, NodeNetworkReadErrors, labels, niface.RxErrors, network.Time.Time)
			writeMetric(writer, NodeNetworkWrite, labels, niface.TxBytes, network.Time.Time)
			writeMetric(writer, NodeNetworkWriteErrors, labels, niface.TxErrors, network.Time.Time)
		}
	}

	for i := range summary.Pods {
		podname := summary.Pods[i].PodRef.Name
		podns := summary.Pods[i].PodRef.Namespace

		commonLabels := map[Label]string{
			KloudliteAccountName: podsMap.PodAccountName(podns, podname),
			KloudliteClusterName: podsMap.PodClusterName(podns, podname),
			KloudliteTrackingId:  podsMap.PodTrackingId(podns, podname),

			PodName:      podname,
			PodNamespace: podns,
		}

		if cpu := summary.Pods[i].CPU; cpu != nil {
			if cpu.UsageNanoCores != nil {
				writeMetric(writer, PodCpuUsage, commonLabels, fn.New(float64(*cpu.UsageNanoCores)/1e6), cpu.Time.Time)
			}
		}

		if mem := summary.Pods[i].Memory; mem != nil {
			writeMetric(writer, PodMemoryUsed, commonLabels, mem.WorkingSetBytes, mem.Time.Time)
			writeMetric(writer, PodMemoryAvailable, commonLabels, mem.AvailableBytes, mem.Time.Time)
		}

		for _, vs := range summary.Pods[i].VolumeStats {
			if pvc := vs.PVCRef; pvc != nil {
				labels := fn.MapMerge(commonLabels, map[Label]string{
					PVCName:      pvc.Name,
					PVCNamespace: pvc.Namespace,
				})
				writeMetric(writer, PodPvcStorageUsed, labels, vs.UsedBytes, vs.Time.Time)
				writeMetric(writer, PodPvcStorageAvail, labels, vs.AvailableBytes, vs.Time.Time)
				writeMetric(writer, PodPvcStorageCapacity, labels, vs.CapacityBytes, vs.Time.Time)
			}
		}

		if network := summary.Pods[i].Network; network != nil {
			for _, niface := range network.Interfaces {
				labels := fn.MapMerge(commonLabels, map[Label]string{
					NetworkInterface: niface.Name,
				})

				writeMetric(writer, PodNetworkRead, labels, niface.RxBytes, network.Time.Time)
				writeMetric(writer, PodNetworkReadErrors, labels, niface.RxErrors, network.Time.Time)
				writeMetric(writer, PodNetworkWrite, labels, niface.TxBytes, network.Time.Time)
				writeMetric(writer, PodNetworkWriteErrors, labels, niface.TxErrors, network.Time.Time)
			}
		}
	}
	return nil
}
