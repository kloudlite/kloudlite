package certs

import (
	"github.com/kloudlite/kloudlite/api/internal/controllers/certs/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Register(mgr ctrl.Manager) error {
	utilruntime.Must(v1.AddToScheme(mgr.GetScheme()))

	caReconciler := CAReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
	if err := caReconciler.SetupWithManager(mgr); err != nil {
		return err
	}

	certReconciler := CertificateReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
	return certReconciler.SetupWithManager(mgr)
}
