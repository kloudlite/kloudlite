package v1

import (
	"fmt"

	fn "github.com/kloudlite/operator/pkg/functions"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var standaloneservicelog = logf.Log.WithName("standaloneservice-resource")

func (r *StandaloneService) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-mongodb-msvc-kloudlite-io-v1-standaloneservice,mutating=true,failurePolicy=fail,sideEffects=None,groups=mongodb.msvc.kloudlite.io,resources=standaloneservices,verbs=create;update,versions=v1,name=mstandaloneservice.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &StandaloneService{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *StandaloneService) Default() {
	standaloneservicelog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
	if r.Spec.OutputSecretName != nil {
		r.Spec.OutputSecretName = fn.New(fmt.Sprintf("msvc-%s", r.Name))
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-mongodb-msvc-kloudlite-io-v1-standaloneservice,mutating=false,failurePolicy=fail,sideEffects=None,groups=mongodb.msvc.kloudlite.io,resources=standaloneservices,verbs=create;update,versions=v1,name=vstandaloneservice.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &StandaloneService{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *StandaloneService) ValidateCreate() error {
	standaloneservicelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *StandaloneService) ValidateUpdate(old runtime.Object) error {
	standaloneservicelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *StandaloneService) ValidateDelete() error {
	standaloneservicelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
