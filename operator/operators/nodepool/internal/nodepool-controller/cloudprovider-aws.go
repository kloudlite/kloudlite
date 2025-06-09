package nodepool_controller

import (
	"encoding/json"
	"fmt"

	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
)

func (r *Reconciler) AwsJobValuesJson(obj *clustersv1.NodePool, nodesMap map[string]clustersv1.NodeProps) (string, error) {
	if obj.Spec.AWS == nil {
		return "", fmt.Errorf(".spec.aws is nil")
	}

	ec2Nodepool := make(map[string]any)
	spotNodepool := make(map[string]any)

	switch obj.Spec.AWS.PoolType {
	case clustersv1.AWSPoolTypeEC2:
		{
			ec2Nodepool = map[string]any{
				"root_volume_type": obj.Spec.AWS.RootVolumeType,
				"root_volume_size": obj.Spec.AWS.RootVolumeSize,
				"instance_type":    obj.Spec.AWS.EC2Pool.InstanceType,
				"ami":              obj.Spec.AWS.EC2Pool.AMI,
				"nodes":            nodesMap,
			}
			spotNodepool = nil
		}
	case clustersv1.AWSPoolTypeSpot:
		{
			if obj.Spec.AWS.SpotPool == nil {
				return "", fmt.Errorf(".spec.aws.spotPool is nil")
			}

			spotNodepool = map[string]any{
				"ami":                          obj.Spec.AWS.SpotPool.AMI,
				"root_volume_type":             obj.Spec.AWS.RootVolumeType,
				"root_volume_size":             obj.Spec.AWS.RootVolumeSize,
				"spot_fleet_tagging_role_name": obj.Spec.AWS.SpotPool.SpotFleetTaggingRoleName,
				"cpu_node": func() map[string]any {
					if obj.Spec.AWS.SpotPool.CpuNode == nil {
						return nil
					}
					return map[string]any{
						"vcpu": map[string]any{
							"min": obj.Spec.AWS.SpotPool.CpuNode.VCpu.Min,
							"max": obj.Spec.AWS.SpotPool.CpuNode.VCpu.Max,
						},
						"memory_per_vcpu": map[string]any{
							"min": obj.Spec.AWS.SpotPool.CpuNode.MemoryPerVCpu.Min,
							"max": obj.Spec.AWS.SpotPool.CpuNode.MemoryPerVCpu.Max,
						},
					}
				}(),
				"gpu_node": func() map[string]any {
					if obj.Spec.AWS.SpotPool.GpuNode == nil {
						return nil
					}

					return map[string]any{
						"instance_types": obj.Spec.AWS.SpotPool.GpuNode.InstanceTypes,
					}
				}(),
				"nodes": nodesMap,
			}
			ec2Nodepool = nil
		}
	}

	variables := map[string]any{
		// INFO: there will be no aws_access_key, aws_secret_key thing, as we expect this autoscaler to run on AWS instances configured with proper IAM instance profile
		// "aws_access_key":             nil,
		// "aws_secret_key":             nil,
		"aws_region":                 obj.Spec.AWS.Region,
		"tracker_id":                 fmt.Sprintf("cluster-%s", r.Env.ClusterName),
		"nodepool_name":              obj.Name,
		"k3s_join_token":             r.Env.K3sJoinToken,
		"k3s_server_public_dns_host": r.Env.K3sServerPublicHost,

		"vpc_id":        obj.Spec.AWS.VPCId,
		"vpc_subnet_id": obj.Spec.AWS.VPCSubnetID,

		"availability_zone":    obj.Spec.AWS.AvailabilityZone,
		"nvidia_gpu_enabled":   obj.Spec.AWS.NvidiaGpuEnabled,
		"iam_instance_profile": obj.Spec.AWS.IAMInstanceProfileRole,

		"ec2_nodepool":  ec2Nodepool,
		"spot_nodepool": spotNodepool,

		"extra_agent_args":     []string{"--snapshotter", "stargz"},
		"save_ssh_key_to_path": "",
		"tags": map[string]string{
			"kloudlite-account": r.Env.AccountName,
			"kloudlite-cluster": r.Env.ClusterName,
		},
		"kloudlite_release":             r.Env.KloudliteRelease,
		"k3s_download_url":              "https://github.com/kloudlite/infrastructure-as-code/releases/download/binaries/k3s",
		"kloudlite_runner_download_url": "https://github.com/kloudlite/infrastructure-as-code/releases/download/binaries/runner-amd64",
	}

	b, err := json.Marshal(variables)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
