package v1

import (
	corev1 "k8s.io/api/core/v1"
	fn "operators.kloudlite.io/lib/functions"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Bool bool

func (b Bool) Status() metav1.ConditionStatus {
	if b {
		return metav1.ConditionTrue
	}
	return metav1.ConditionFalse
}

type Condition struct {
	Type               string
	Status             string // "True", "False", "Unknown"
	ObservedGeneration int64
	Reason             string
	Message            string
}

type Operations struct {
	Apply  string `json:"create"`
	Delete string `json:"delete"`
}

type Output struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type ResourceType string

const (
	SecretType ResourceType = "secret"
	ConfigType ResourceType = "config"
)

func ParseVolumes(containers []AppContainer) (volumes []corev1.Volume, volumeMounts [][]corev1.VolumeMount) {
	m := map[string][]ContainerVolume{}

	for _, container := range containers {
		var mounts []corev1.VolumeMount
		for _, volume := range container.Volumes {
			volName := fn.Md5([]byte(volume.MountPath))
			if len(volName) > 50 {
				volName = volName[:50]
			}
			mounts = append(
				mounts, corev1.VolumeMount{
					Name:      volName,
					MountPath: volume.MountPath,
				},
			)
			if m[volName] == nil {
				m[volName] = []ContainerVolume{}
			}
			m[volName] = append(m[volName], volume)
		}

		volumeMounts = append(volumeMounts, mounts)
	}

	for k, cVolumes := range m {
		volume := corev1.Volume{Name: k}

		// len == 1, without projection
		if len(cVolumes) == 1 {
			volm := cVolumes[0]

			var kp []corev1.KeyToPath
			if len(volm.Items) > 0 {
				for _, item := range volm.Items {
					kp = append(
						kp, corev1.KeyToPath{
							Key:  item.Key,
							Path: item.FileName,
							Mode: nil,
						},
					)
				}
			}

			switch volm.Type {
			case SecretType:
				{
					volume.VolumeSource.Secret = &corev1.SecretVolumeSource{
						SecretName: volm.RefName,
						Items:      kp,
					}
				}
			case ConfigType:
				{
					volume.VolumeSource.ConfigMap = &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: volm.RefName,
						},
						Items: kp,
					}
				}
			}
		}

		// len > 1, with projection
		if len(cVolumes) > 1 {
			volume.VolumeSource.Projected = &corev1.ProjectedVolumeSource{}
			for _, volm := range cVolumes {
				projection := corev1.VolumeProjection{}
				var kp []corev1.KeyToPath
				if len(volm.Items) > 0 {
					for _, item := range volm.Items {
						kp = append(
							kp, corev1.KeyToPath{
								Key:  item.Key,
								Path: item.FileName,
								Mode: nil,
							},
						)
					}
				}
				switch volm.Type {
				case SecretType:
					{
						projection.Secret = &corev1.SecretProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: volm.RefName,
							},
							Items: kp,
						}
					}
				case ConfigType:
					{
						projection.ConfigMap = &corev1.ConfigMapProjection{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: volm.RefName,
							},
							Items: kp,
						}
					}
				}
			}
		}
		volumes = append(volumes, volume)
	}

	return volumes, volumeMounts
}
