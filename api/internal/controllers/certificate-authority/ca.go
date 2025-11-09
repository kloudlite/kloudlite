package certificateauthority

import (
	"github.com/kloudlite/kloudlite/api/internal/controllers/certificate-authority/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Register(mgr ctrl.Manager) error {
	utilruntime.Must(v1.AddToScheme(mgr.GetScheme()))

	reconciler := Reconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}

	return reconciler.SetupWithManager(mgr)
}
