package v1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var AppDeletion = struct {
	DeploymentDeleted string
	ServiceDeleted    string
}{
	"deployment-deleted",
	"service deleted",
}

func (app *App) ExpectedConditions() []string {
	if app.GetDeletionTimestamp() != nil {
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
	}
}

const TypeReady = "ready"

func (app *App) HasConditionsMet() bool {
	b := true
	for _, item := range app.ExpectedConditions() {
		b = b && meta.IsStatusConditionTrue(app.Status.Conditions, item)
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
