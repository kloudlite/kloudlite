package kloudlite

type MetricName string

const (
	NodeMemoryUsed      MetricName = "kl_node_mem_used"
	NodeMemoryAvailable MetricName = "kl_node_mem_avail"

	NodeCpuUtilization MetricName = "kl_node_cpu_utilization"
	NodeCpuUsage       MetricName = "kl_node_cpu_usage"

	NodeStorageUsed     MetricName = "kl_node_storage_used"
	NodeStorageAvail    MetricName = "kl_node_storage_avail"
	NodeStorageCapacity MetricName = "kl_node_storage_capacity"

	NodeNetworkRead        MetricName = "kl_node_network_read"
	NodeNetworkReadErrors  MetricName = "kl_node_network_read_errors"
	NodeNetworkWrite       MetricName = "kl_node_network_write"
	NodeNetworkWriteErrors MetricName = "kl_node_network_write_errors"

	PodMemoryUsed      MetricName = "kl_pod_mem_used"
	PodMemoryAvailable MetricName = "kl_pod_mem_avail"

	PodCpuUsage MetricName = "kl_pod_cpu_usage"

	PodPvcStorageUsed     MetricName = "kl_pod_pvc_storage_used"
	PodPvcStorageAvail    MetricName = "kl_pod_pvc_storage_avail"
	PodPvcStorageCapacity MetricName = "kl_pod_pvc_storage_capacity"

	PodNetworkRead        MetricName = "kl_pod_network_read"
	PodNetworkReadErrors  MetricName = "kl_pod_network_read_errors"
	PodNetworkWrite       MetricName = "kl_pod_network_write"
	PodNetworkWriteErrors MetricName = "kl_pod_network_write_errors"
)

type Label string

const (
	KloudliteAccountName Label = "kl_account_name"
	KloudliteClusterName Label = "kl_cluster_name"
	KloudliteTrackingId  Label = "kl_tracking_id"

	// KloudliteRecordVersion Label = "kl_record_version"

	NodeName Label = "node_name"

	NetworkInterface Label = "iface"

	PodNamespace Label = "pod_ns"
	PodName      Label = "pod_name"

	PVCNamespace Label = "pvc_ns"
	PVCName      Label = "pvc_name"
)

var LabelSortingOrder = map[Label]int{
	KloudliteAccountName: 1,
	KloudliteClusterName: 2,
	KloudliteTrackingId:  3,

	NodeName:         4,
	NetworkInterface: 5,

	PodNamespace: 11,
	PodName:      12,

	PVCName:      13,
	PVCNamespace: 14,
}
