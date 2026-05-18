package machinetypes

import (
	"fmt"
	"strings"

	workmachinev1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Defaults(provider string) ([]workmachinev1.MachineType, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "aws":
		return clone(defaultAWS), nil
	case "gcp":
		return clone(defaultGCP), nil
	case "azure":
		return clone(defaultAzure), nil
	case "oci":
		return clone(defaultOCI), nil
	default:
		return nil, fmt.Errorf("unsupported cloud provider %q", provider)
	}
}

func clone(items []workmachinev1.MachineType) []workmachinev1.MachineType {
	out := make([]workmachinev1.MachineType, len(items))
	copy(out, items)
	return out
}

func machineType(name, displayName, description, category, cpu, memory, gpu string, isDefault bool, priority int32) workmachinev1.MachineType {
	return workmachinev1.MachineType{
		TypeMeta: metav1.TypeMeta{APIVersion: workmachinev1.GroupVersion.String(), Kind: "MachineType"},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: workmachinev1.MachineTypeSpec{
			DisplayName: displayName,
			Description: description,
			Category:    category,
			Resources: workmachinev1.MachineResources{
				CPU:    cpu,
				Memory: memory,
				GPU:    gpu,
			},
			Active:    true,
			IsDefault: isDefault,
			Priority:  priority,
		},
	}
}

var defaultAWS = []workmachinev1.MachineType{
	machineType("aws-t3-medium", "AWS t3.medium", "General purpose development machine", "development", "2", "4Gi", "", true, 10),
	machineType("aws-t3-large", "AWS t3.large", "Larger general purpose machine", "general", "2", "8Gi", "", false, 20),
	machineType("aws-m5-xlarge", "AWS m5.xlarge", "Compute and memory balanced machine", "general", "4", "16Gi", "", false, 30),
	machineType("aws-g4dn-xlarge", "AWS g4dn.xlarge", "GPU machine", "gpu", "4", "16Gi", "1", false, 40),
}

var defaultGCP = []workmachinev1.MachineType{
	machineType("gcp-e2-standard-2", "GCP e2-standard-2", "General purpose development machine", "development", "2", "8Gi", "", true, 10),
	machineType("gcp-e2-standard-4", "GCP e2-standard-4", "Larger general purpose machine", "general", "4", "16Gi", "", false, 20),
	machineType("gcp-n2-standard-8", "GCP n2-standard-8", "Compute optimized machine", "compute-optimized", "8", "32Gi", "", false, 30),
	machineType("gcp-g2-standard-4", "GCP g2-standard-4", "GPU machine", "gpu", "4", "16Gi", "1", false, 40),
}

var defaultAzure = []workmachinev1.MachineType{
	machineType("azure-standard-b2s", "Azure Standard_B2s", "General purpose development machine", "development", "2", "4Gi", "", true, 10),
	machineType("azure-standard-d4s-v5", "Azure Standard_D4s_v5", "Larger general purpose machine", "general", "4", "16Gi", "", false, 20),
	machineType("azure-standard-f8s-v2", "Azure Standard_F8s_v2", "Compute optimized machine", "compute-optimized", "8", "16Gi", "", false, 30),
	machineType("azure-standard-nc4as-t4-v3", "Azure Standard_NC4as_T4_v3", "GPU machine", "gpu", "4", "28Gi", "1", false, 40),
}

var defaultOCI = []workmachinev1.MachineType{
	machineType("oci-vm-standard-e4-flex-small", "OCI VM.Standard.E4.Flex Small", "General purpose development machine", "development", "2", "8Gi", "", true, 10),
	machineType("oci-vm-standard-e4-flex-medium", "OCI VM.Standard.E4.Flex Medium", "Larger general purpose machine", "general", "4", "16Gi", "", false, 20),
	machineType("oci-vm-optimized3-flex", "OCI VM.Optimized3.Flex", "Compute optimized machine", "compute-optimized", "8", "32Gi", "", false, 30),
	machineType("oci-vm-gpu-a10-1", "OCI VM.GPU.A10.1", "GPU machine", "gpu", "15", "240Gi", "1", false, 40),
}
