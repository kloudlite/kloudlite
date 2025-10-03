package controllers

import (
	"fmt"
	"strconv"

	composego "github.com/compose-spec/compose-go/v2/types"
	environmentsv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/environments/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// ComposeResources holds all Kubernetes resources converted from docker-compose
type ComposeResources struct {
	Deployments  []*appsv1.Deployment
	Services     []*corev1.Service
	ConfigMaps   []*corev1.ConfigMap
	Secrets      []*corev1.Secret
	PVCs         []*corev1.PersistentVolumeClaim
	ServiceNames []string
}

// ConvertComposeToK8s converts a docker-compose project to Kubernetes resources
func ConvertComposeToK8s(
	project *composego.Project,
	composition *environmentsv1.Composition,
	namespace string,
) (*ComposeResources, error) {
	resources := &ComposeResources{
		Deployments:  make([]*appsv1.Deployment, 0),
		Services:     make([]*corev1.Service, 0),
		ConfigMaps:   make([]*corev1.ConfigMap, 0),
		Secrets:      make([]*corev1.Secret, 0),
		PVCs:         make([]*corev1.PersistentVolumeClaim, 0),
		ServiceNames: make([]string, 0),
	}

	// Common labels for all resources
	commonLabels := map[string]string{
		"kloudlite.io/docker-composition": composition.Name,
		"kloudlite.io/managed":             "true",
	}

	// Convert volumes first (they need to exist before services)
	for volumeName, volume := range project.Volumes {
		pvc := convertVolumeToPVC(volumeName, volume, composition, namespace, commonLabels)
		resources.PVCs = append(resources.PVCs, pvc)
	}

	// Convert each service
	for serviceName, service := range project.Services {
		resources.ServiceNames = append(resources.ServiceNames, serviceName)

		// Create Deployment
		deployment, err := convertServiceToDeployment(
			serviceName,
			service,
			composition,
			namespace,
			commonLabels,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to convert service %s: %w", serviceName, err)
		}
		resources.Deployments = append(resources.Deployments, deployment)

		// Create Service (if ports are exposed)
		if len(service.Ports) > 0 {
			k8sService := convertServiceToK8sService(
				serviceName,
				service,
				composition,
				namespace,
				commonLabels,
			)
			resources.Services = append(resources.Services, k8sService)
		}
	}

	return resources, nil
}

// convertServiceToDeployment converts a docker-compose service to a Kubernetes Deployment
func convertServiceToDeployment(
	serviceName string,
	service composego.ServiceConfig,
	composition *environmentsv1.Composition,
	namespace string,
	commonLabels map[string]string,
) (*appsv1.Deployment, error) {
	// Service-specific labels
	labels := make(map[string]string)
	for k, v := range commonLabels {
		labels[k] = v
	}
	labels["kloudlite.io/service"] = serviceName

	// Determine replicas - default to 1
	replicas := int32(1)

	// Use service-level replicas from docker-compose deploy section
	if service.Deploy != nil && service.Deploy.Replicas != nil {
		replicas = int32(*service.Deploy.Replicas)
	}

	// Check for resource overrides (highest priority)
	if override, ok := composition.Spec.ResourceOverrides[serviceName]; ok {
		if override.Replicas != nil {
			replicas = *override.Replicas
		}
	}

	// Build container
	container := corev1.Container{
		Name:  serviceName,
		Image: service.Image,
	}

	// Add command and args if specified
	if len(service.Command) > 0 {
		container.Command = service.Command
	}
	if len(service.Entrypoint) > 0 {
		container.Command = service.Entrypoint
	}

	// Add environment variables
	envVars := make([]corev1.EnvVar, 0)
	for key, val := range service.Environment {
		if val != nil {
			envVars = append(envVars, corev1.EnvVar{
				Name:  key,
				Value: *val,
			})
		}
	}
	// Add composition-level env vars
	for key, val := range composition.Spec.EnvVars {
		envVars = append(envVars, corev1.EnvVar{
			Name:  key,
			Value: val,
		})
	}
	container.Env = envVars

	// Add ports
	containerPorts := make([]corev1.ContainerPort, 0)
	for _, port := range service.Ports {
		if port.Target != 0 {
			containerPorts = append(containerPorts, corev1.ContainerPort{
				ContainerPort: int32(port.Target),
				Protocol:      corev1.ProtocolTCP,
			})
		}
	}
	container.Ports = containerPorts

	// Add resource limits
	resources := corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}

	if service.Deploy != nil && service.Deploy.Resources.Limits != nil {
		if service.Deploy.Resources.Limits.NanoCPUs > 0 {
			cpuLimit := convertCPU(float64(service.Deploy.Resources.Limits.NanoCPUs))
			resources.Limits[corev1.ResourceCPU] = resource.MustParse(cpuLimit)
		}
		if service.Deploy.Resources.Limits.MemoryBytes != 0 {
			resources.Limits[corev1.ResourceMemory] = *resource.NewQuantity(
				int64(service.Deploy.Resources.Limits.MemoryBytes),
				resource.BinarySI,
			)
		}
	}

	// Apply resource overrides
	if override, ok := composition.Spec.ResourceOverrides[serviceName]; ok {
		if override.CPU != "" {
			resources.Limits[corev1.ResourceCPU] = resource.MustParse(override.CPU)
		}
		if override.Memory != "" {
			resources.Limits[corev1.ResourceMemory] = resource.MustParse(override.Memory)
		}
	}

	container.Resources = resources

	// Add volume mounts
	volumeMounts := make([]corev1.VolumeMount, 0)
	volumes := make([]corev1.Volume, 0)

	for _, vol := range service.Volumes {
		if vol.Type == "volume" && vol.Source != "" {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      vol.Source,
				MountPath: vol.Target,
			})
			volumes = append(volumes, corev1.Volume{
				Name: vol.Source,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: fmt.Sprintf("%s-%s", composition.Name, vol.Source),
					},
				},
			})
		}
	}
	container.VolumeMounts = volumeMounts

	// Create deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", composition.Name, serviceName),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
					Volumes:    volumes,
				},
			},
		},
	}

	return deployment, nil
}

// convertServiceToK8sService converts docker-compose service ports to Kubernetes Service
func convertServiceToK8sService(
	serviceName string,
	service composego.ServiceConfig,
	composition *environmentsv1.Composition,
	namespace string,
	commonLabels map[string]string,
) *corev1.Service {
	labels := make(map[string]string)
	for k, v := range commonLabels {
		labels[k] = v
	}
	labels["kloudlite.io/service"] = serviceName

	ports := make([]corev1.ServicePort, 0)
	for i, port := range service.Ports {
		publishedPort := port.Target
		if port.Published != "" {
			// Parse Published port string to uint32
			pubInt, err := strconv.ParseUint(port.Published, 10, 32)
			if err == nil {
				publishedPort = uint32(pubInt)
			}
		}

		servicePort := corev1.ServicePort{
			Name:       fmt.Sprintf("port-%d", i),
			Port:       int32(publishedPort),
			TargetPort: intstr.FromInt(int(port.Target)),
			Protocol:   corev1.ProtocolTCP,
		}
		ports = append(ports, servicePort)
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", composition.Name, serviceName),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports:    ports,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
}

// convertVolumeToPVC converts docker-compose volume to PersistentVolumeClaim
func convertVolumeToPVC(
	volumeName string,
	volume composego.VolumeConfig,
	composition *environmentsv1.Composition,
	namespace string,
	commonLabels map[string]string,
) *corev1.PersistentVolumeClaim {
	labels := make(map[string]string)
	for k, v := range commonLabels {
		labels[k] = v
	}
	labels["kloudlite.io/volume"] = volumeName

	// Default size
	size := resource.MustParse("1Gi")

	// TODO: Parse size from driver_opts if specified

	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", composition.Name, volumeName),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: size,
				},
			},
		},
	}
}

// convertCPU converts docker-compose CPU format to Kubernetes format
func convertCPU(nanoCPUs float64) string {
	// NanoCPUs format: 0.5 means 0.5 CPU cores
	// Kubernetes format: "500m" means 500 millicores (0.5 cores)
	if nanoCPUs == 0 {
		return "1"
	}

	// Convert to millicores
	millicores := int(nanoCPUs * 1000)
	return fmt.Sprintf("%dm", millicores)
}
