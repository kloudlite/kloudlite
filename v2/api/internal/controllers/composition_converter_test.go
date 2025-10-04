package controllers

import (
	"testing"

	composego "github.com/compose-spec/compose-go/v2/types"
	environmentsv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/environments/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertServiceToDeployment_WithFilesVolumes(t *testing.T) {
	service := composego.ServiceConfig{
		Name:  "web",
		Image: "nginx:latest",
		Volumes: []composego.ServiceVolumeConfig{
			{
				Type:   "bind",
				Source: "/files/app.yml",
				Target: "/etc/nginx/nginx.conf",
			},
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
	}

	envData := &EnvironmentData{
		EnvVars: make(map[string]string),
		Secrets: make(map[string]string),
		ConfigFiles: map[string]string{
			"app.yml": "nginx config content",
		},
	}

	commonLabels := map[string]string{
		"app": "test",
	}

	deployment, err := convertServiceToDeployment("web", service, composition, "test-ns", commonLabels, envData)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	// Check volumes
	assert.Equal(t, 1, len(deployment.Spec.Template.Spec.Volumes))
	volume := deployment.Spec.Template.Spec.Volumes[0]
	assert.Equal(t, "env-file-app-yml", volume.Name) // Dots replaced with dashes
	assert.NotNil(t, volume.ConfigMap)
	assert.Equal(t, "env-file-app.yml", volume.ConfigMap.Name)

	// Check volume mounts
	assert.Equal(t, 1, len(deployment.Spec.Template.Spec.Containers))
	container := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, 1, len(container.VolumeMounts))
	volumeMount := container.VolumeMounts[0]
	assert.Equal(t, "env-file-app-yml", volumeMount.Name)
	assert.Equal(t, "/etc/nginx/nginx.conf", volumeMount.MountPath)
	assert.Equal(t, "app.yml", volumeMount.SubPath)
}

func TestConvertServiceToDeployment_WithFilesVolumes_DotInFilename(t *testing.T) {
	service := composego.ServiceConfig{
		Name:  "web",
		Image: "nginx:latest",
		Volumes: []composego.ServiceVolumeConfig{
			{
				Type:   "bind",
				Source: "/files/config.json",
				Target: "/app/config.json",
			},
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
	}

	envData := &EnvironmentData{
		EnvVars: make(map[string]string),
		Secrets: make(map[string]string),
		ConfigFiles: map[string]string{
			"config.json": "{}",
		},
	}

	commonLabels := map[string]string{}

	deployment, err := convertServiceToDeployment("web", service, composition, "test-ns", commonLabels, envData)
	assert.NoError(t, err)

	// Volume name should have dots replaced with dashes
	volume := deployment.Spec.Template.Spec.Volumes[0]
	assert.Equal(t, "env-file-config-json", volume.Name)

	// ConfigMap name should keep the original filename
	assert.Equal(t, "env-file-config.json", volume.ConfigMap.Name)

	// SubPath should use original filename
	volumeMount := deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0]
	assert.Equal(t, "config.json", volumeMount.SubPath)
}

func TestConvertServiceToDeployment_WithFilesVolumes_MissingConfigFile(t *testing.T) {
	service := composego.ServiceConfig{
		Name:  "web",
		Image: "nginx:latest",
		Volumes: []composego.ServiceVolumeConfig{
			{
				Type:   "bind",
				Source: "/files/app.yml",
				Target: "/etc/nginx/nginx.conf",
			},
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
	}

	// ConfigFile "app.yml" does NOT exist
	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	commonLabels := map[string]string{}

	deployment, err := convertServiceToDeployment("web", service, composition, "test-ns", commonLabels, envData)
	assert.NoError(t, err)

	// Volume should NOT be created if ConfigFile doesn't exist
	assert.Equal(t, 0, len(deployment.Spec.Template.Spec.Volumes))
	assert.Equal(t, 0, len(deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
}

func TestConvertServiceToDeployment_MixedVolumes(t *testing.T) {
	service := composego.ServiceConfig{
		Name:  "web",
		Image: "postgres:latest",
		Volumes: []composego.ServiceVolumeConfig{
			{
				Type:   "bind",
				Source: "/files/postgresql.conf",
				Target: "/etc/postgresql/postgresql.conf",
			},
			{
				Type:   "volume",
				Source: "data",
				Target: "/var/lib/postgresql/data",
			},
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
	}

	envData := &EnvironmentData{
		EnvVars: make(map[string]string),
		Secrets: make(map[string]string),
		ConfigFiles: map[string]string{
			"postgresql.conf": "config content",
		},
	}

	commonLabels := map[string]string{}

	deployment, err := convertServiceToDeployment("web", service, composition, "test-ns", commonLabels, envData)
	assert.NoError(t, err)

	// Should have 2 volumes: ConfigMap + PVC
	assert.Equal(t, 2, len(deployment.Spec.Template.Spec.Volumes))

	// Check ConfigMap volume
	var configMapVolume *corev1.Volume
	var pvcVolume *corev1.Volume
	for i := range deployment.Spec.Template.Spec.Volumes {
		vol := &deployment.Spec.Template.Spec.Volumes[i]
		if vol.ConfigMap != nil {
			configMapVolume = vol
		}
		if vol.PersistentVolumeClaim != nil {
			pvcVolume = vol
		}
	}

	assert.NotNil(t, configMapVolume)
	assert.Equal(t, "env-file-postgresql-conf", configMapVolume.Name)

	assert.NotNil(t, pvcVolume)
	assert.Equal(t, "data", pvcVolume.Name)
	assert.Equal(t, "test-comp-data", pvcVolume.PersistentVolumeClaim.ClaimName)

	// Check volume mounts
	assert.Equal(t, 2, len(deployment.Spec.Template.Spec.Containers[0].VolumeMounts))
}

func TestConvertComposeToK8s_WithFilesVolumes(t *testing.T) {
	project := &composego.Project{
		Name: "test-project",
		Services: composego.Services{
			"web": composego.ServiceConfig{
				Name:  "web",
				Image: "nginx:latest",
				Volumes: []composego.ServiceVolumeConfig{
					{
						Type:   "bind",
						Source: "/files/nginx.conf",
						Target: "/etc/nginx/nginx.conf",
					},
				},
			},
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
	}

	envData := &EnvironmentData{
		EnvVars: make(map[string]string),
		Secrets: make(map[string]string),
		ConfigFiles: map[string]string{
			"nginx.conf": "server { ... }",
		},
	}

	resources, err := ConvertComposeToK8s(project, composition, "test-ns", envData)
	assert.NoError(t, err)
	assert.NotNil(t, resources)

	// Should have 1 deployment
	assert.Equal(t, 1, len(resources.Deployments))
	deployment := resources.Deployments[0]

	// Verify ConfigMap volume was created
	assert.Equal(t, 1, len(deployment.Spec.Template.Spec.Volumes))
	volume := deployment.Spec.Template.Spec.Volumes[0]
	assert.NotNil(t, volume.ConfigMap)
	assert.Equal(t, "env-file-nginx.conf", volume.ConfigMap.Name)
}

func TestConvertCPU(t *testing.T) {
	// Test zero CPU (should return default "1")
	assert.Equal(t, "1", convertCPU(0))

	// Test fractional CPU values
	assert.Equal(t, "500m", convertCPU(0.5))
	assert.Equal(t, "250m", convertCPU(0.25))
	assert.Equal(t, "750m", convertCPU(0.75))

	// Test whole number CPU values
	assert.Equal(t, "1000m", convertCPU(1.0))
	assert.Equal(t, "2000m", convertCPU(2.0))
	assert.Equal(t, "4000m", convertCPU(4.0))

	// Test high precision values
	assert.Equal(t, "333m", convertCPU(0.333))
	assert.Equal(t, "1500m", convertCPU(1.5))
}

func TestConvertVolumeToPVC(t *testing.T) {
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
	}

	commonLabels := map[string]string{
		"kloudlite.io/docker-composition": "test-comp",
		"kloudlite.io/managed":            "true",
	}

	volume := composego.VolumeConfig{
		Name: "data",
	}

	pvc := convertVolumeToPVC("data", volume, composition, "test-ns", commonLabels)

	// Verify PVC metadata
	assert.Equal(t, "test-comp-data", pvc.Name)
	assert.Equal(t, "test-ns", pvc.Namespace)

	// Verify labels
	assert.Equal(t, "test-comp", pvc.Labels["kloudlite.io/docker-composition"])
	assert.Equal(t, "true", pvc.Labels["kloudlite.io/managed"])
	assert.Equal(t, "data", pvc.Labels["kloudlite.io/volume"])

	// Verify PVC spec
	assert.Equal(t, 1, len(pvc.Spec.AccessModes))
	assert.Equal(t, corev1.ReadWriteOnce, pvc.Spec.AccessModes[0])

	// Verify default size (1Gi)
	storageSize := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	assert.Equal(t, "1Gi", storageSize.String())
}

func TestConvertServiceToDeployment_WithResourceLimits(t *testing.T) {
	service := composego.ServiceConfig{
		Name:  "app",
		Image: "myapp:latest",
		Deploy: &composego.DeployConfig{
			Resources: composego.Resources{
				Limits: &composego.Resource{
					NanoCPUs:    500000000,  // 0.5 CPU
					MemoryBytes: 536870912,  // 512Mi
				},
			},
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
	}

	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	commonLabels := map[string]string{}

	deployment, err := convertServiceToDeployment("app", service, composition, "test-ns", commonLabels, envData)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	// Verify CPU limit
	container := deployment.Spec.Template.Spec.Containers[0]
	cpuLimit := container.Resources.Limits[corev1.ResourceCPU]
	// Kubernetes uses uppercase M for millicores
	assert.Equal(t, "500M", cpuLimit.String())

	// Verify Memory limit
	memLimit := container.Resources.Limits[corev1.ResourceMemory]
	assert.Equal(t, int64(536870912), memLimit.Value())
}

func TestConvertServiceToDeployment_WithResourceOverrides(t *testing.T) {
	service := composego.ServiceConfig{
		Name:  "app",
		Image: "myapp:latest",
		Deploy: &composego.DeployConfig{
			Resources: composego.Resources{
				Limits: &composego.Resource{
					NanoCPUs:    500000000, // 0.5 CPU (will be overridden)
					MemoryBytes: 536870912, // 512Mi (will be overridden)
				},
			},
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
		Spec: environmentsv1.CompositionSpec{
			ResourceOverrides: map[string]environmentsv1.ServiceResourceOverride{
				"app": {
					CPU:    "2",
					Memory: "2Gi",
				},
			},
		},
	}

	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	commonLabels := map[string]string{}

	deployment, err := convertServiceToDeployment("app", service, composition, "test-ns", commonLabels, envData)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	// Verify overridden CPU limit
	container := deployment.Spec.Template.Spec.Containers[0]
	cpuLimit := container.Resources.Limits[corev1.ResourceCPU]
	assert.Equal(t, "2", cpuLimit.String())

	// Verify overridden Memory limit
	memLimit := container.Resources.Limits[corev1.ResourceMemory]
	assert.Equal(t, "2Gi", memLimit.String())
}

func TestConvertServiceToDeployment_WithReplicas(t *testing.T) {
	replicas := 3
	service := composego.ServiceConfig{
		Name:  "app",
		Image: "myapp:latest",
		Deploy: &composego.DeployConfig{
			Replicas: &replicas,
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
	}

	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	commonLabels := map[string]string{}

	deployment, err := convertServiceToDeployment("app", service, composition, "test-ns", commonLabels, envData)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	// Verify replicas from deploy section
	assert.Equal(t, int32(3), *deployment.Spec.Replicas)
}

func TestConvertServiceToDeployment_WithReplicasOverride(t *testing.T) {
	deployReplicas := 3
	overrideReplicas := int32(5)

	service := composego.ServiceConfig{
		Name:  "app",
		Image: "myapp:latest",
		Deploy: &composego.DeployConfig{
			Replicas: &deployReplicas,
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
		Spec: environmentsv1.CompositionSpec{
			ResourceOverrides: map[string]environmentsv1.ServiceResourceOverride{
				"app": {
					Replicas: &overrideReplicas,
				},
			},
		},
	}

	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	commonLabels := map[string]string{}

	deployment, err := convertServiceToDeployment("app", service, composition, "test-ns", commonLabels, envData)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	// Verify override takes precedence
	assert.Equal(t, int32(5), *deployment.Spec.Replicas)
}

func TestConvertServiceToDeployment_WithCompositionEnvVars(t *testing.T) {
	apiURL := "https://api.example.com"
	service := composego.ServiceConfig{
		Name:  "app",
		Image: "myapp:latest",
		Environment: composego.MappingWithEquals{
			"SERVICE_VAR": &apiURL,
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
		Spec: environmentsv1.CompositionSpec{
			EnvVars: map[string]string{
				"GLOBAL_VAR": "global-value",
				"REGION":     "us-west-2",
			},
		},
	}

	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	commonLabels := map[string]string{}

	deployment, err := convertServiceToDeployment("app", service, composition, "test-ns", commonLabels, envData)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	// Verify environment variables
	container := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, 3, len(container.Env))

	// Find and verify each env var
	envMap := make(map[string]string)
	for _, env := range container.Env {
		envMap[env.Name] = env.Value
	}

	assert.Equal(t, "https://api.example.com", envMap["SERVICE_VAR"])
	assert.Equal(t, "global-value", envMap["GLOBAL_VAR"])
	assert.Equal(t, "us-west-2", envMap["REGION"])
}

func TestConvertServiceToDeployment_WithCommandAndEntrypoint(t *testing.T) {
	service := composego.ServiceConfig{
		Name:       "app",
		Image:      "myapp:latest",
		Command:    []string{"serve", "--port=8080"},
		Entrypoint: []string{"/bin/custom-entrypoint.sh"},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
	}

	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	commonLabels := map[string]string{}

	deployment, err := convertServiceToDeployment("app", service, composition, "test-ns", commonLabels, envData)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	// Verify entrypoint takes precedence over command
	container := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, []string{"/bin/custom-entrypoint.sh"}, container.Command)
}

func TestConvertComposeToK8s_WithNamedVolumes(t *testing.T) {
	project := &composego.Project{
		Name: "test-project",
		Services: composego.Services{
			"db": composego.ServiceConfig{
				Name:  "db",
				Image: "postgres:latest",
				Volumes: []composego.ServiceVolumeConfig{
					{
						Type:   "volume",
						Source: "pgdata",
						Target: "/var/lib/postgresql/data",
					},
				},
			},
		},
		Volumes: composego.Volumes{
			"pgdata": composego.VolumeConfig{
				Name: "pgdata",
			},
		},
	}

	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-comp",
			Namespace: "test-ns",
		},
	}

	envData := &EnvironmentData{
		EnvVars:     make(map[string]string),
		Secrets:     make(map[string]string),
		ConfigFiles: make(map[string]string),
	}

	resources, err := ConvertComposeToK8s(project, composition, "test-ns", envData)
	assert.NoError(t, err)
	assert.NotNil(t, resources)

	// Should have 1 PVC
	assert.Equal(t, 1, len(resources.PVCs))
	pvc := resources.PVCs[0]
	assert.Equal(t, "test-comp-pgdata", pvc.Name)
	assert.Equal(t, "test-ns", pvc.Namespace)

	// Should have 1 deployment with volume mount
	assert.Equal(t, 1, len(resources.Deployments))
	deployment := resources.Deployments[0]
	assert.Equal(t, 1, len(deployment.Spec.Template.Spec.Volumes))

	volume := deployment.Spec.Template.Spec.Volumes[0]
	assert.Equal(t, "pgdata", volume.Name)
	assert.NotNil(t, volume.PersistentVolumeClaim)
	assert.Equal(t, "test-comp-pgdata", volume.PersistentVolumeClaim.ClaimName)

	// Verify volume mount
	container := deployment.Spec.Template.Spec.Containers[0]
	assert.Equal(t, 1, len(container.VolumeMounts))
	volumeMount := container.VolumeMounts[0]
	assert.Equal(t, "pgdata", volumeMount.Name)
	assert.Equal(t, "/var/lib/postgresql/data", volumeMount.MountPath)
}
