package nodepool_controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobVars struct {
	JobName         string
	JobNamespace    string
	JobNodeSelector map[string]string

	IACJobImage string
	Labels      map[string]string
	Annotations map[string]string

	OwnerRefs []metav1.OwnerReference

	NodepoolName           string
	TfStateSecretNamespace string

	ValuesJson string
}
