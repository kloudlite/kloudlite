package conditions

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ApplicationDeleted = "application-deleted"
const ConfigMapsDeleted = "configmaps-deleted"
const SecretsDeleted = "secrets-deleted"
const RoutersDeleted = "routers-deleted"
const NamespaceDeleted = "namespace-deleted"

type Condition struct {
	metav1.Condition
	typePrrefix string
}

func ToMap(c []metav1.Condition) map[string]metav1.Condition {
	m := make(map[string]metav1.Condition, len(c))
	for _, condition := range c {
		m[condition.Type] = condition
	}
	return m
}

func init() {
	var c []Condition
	meta.SetStatusCondition(&c.map(item=>item.Condition), Condition{
		Type: ApplicationDeleted,
		Status: metav1.ConditionTrue,
		Reason: "Deleted",
		Message: "Application has been deleted",
	})
	if err := Set(
		c, metav1.Condition{
			Type:   ApplicationDeleted,
			Status: metav1.ConditionTrue,
		},
	); err != nil {
		return
	}
}
