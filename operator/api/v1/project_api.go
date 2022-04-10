package v1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ProjectDeletion = struct {
	ApplicationDeleted string
	ConfigMapsDeleted  string
	SecretsDeleted     string
	RoutersDeleted     string
	NamespaceDeleted   string
	VolumesDeleted     string
	ManagedSvcDeleted  string
	ManagedResDeleted  string
}{
	"application-deleted",
	"configmaps-deleted",
	"secrets-deleted",
	"routers-deleted",
	"namespace-deleted",
	"volumes-deleted",
	"msvc-deleted",
	"mres-deleted",
}

func (p *Project) ExpectedConditions() []string {
	if p.GetDeletionTimestamp() != nil {
		return []string{
			ProjectDeletion.ConfigMapsDeleted,
			ProjectDeletion.SecretsDeleted,
			ProjectDeletion.RoutersDeleted,
			ProjectDeletion.ApplicationDeleted,
			ProjectDeletion.ManagedResDeleted,
			ProjectDeletion.ManagedSvcDeleted,
			ProjectDeletion.VolumesDeleted,
			ProjectDeletion.NamespaceDeleted,
		}
	}

	return []string{
		"namespace-exists",
	}
}

const TypeReady = "ready"

func (p *Project) HasConditionsMet() bool {
	b := true
	for _, item := range p.ExpectedConditions() {
		b = b && meta.IsStatusConditionTrue(p.Status.Conditions, item)
		if !b {
			return false
		}
	}
	return b
}

func (p *Project) IsNextGeneration() bool {
	readyCond := meta.FindStatusCondition(p.Status.Conditions, TypeReady)
	if readyCond == nil {
		return false
	}
	return readyCond.ObservedGeneration != p.Generation
}

func (p *Project) IsReady() bool {
	readyCond := meta.FindStatusCondition(p.Status.Conditions, TypeReady)
	if readyCond == nil {
		return true
	}
	return readyCond.Status == metav1.ConditionTrue
}

func (p *Project) SetReady() bool {
	p.Status.Conditions = []metav1.Condition{}
	meta.SetStatusCondition(&p.Status.Conditions, metav1.Condition{
		Type:               TypeReady,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: p.Generation,
		Reason:             "operator",
		Message:            "",
	})
	return true
}
