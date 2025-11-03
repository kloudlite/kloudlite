package workmachine

import (
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/kubectl"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Register(mgr ctrl.Manager) error {
	utilruntime.Must(v1.AddToScheme(mgr.GetScheme()))

	yamlClient, err := kubectl.NewYAMLClient(mgr.GetConfig(), kubectl.YAMLClientOpts{})
	if err != nil {
		return err
	}

	reconciler := WorkMachineReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),

		YAMLClient: yamlClient,
	}

	return reconciler.SetupWithManager(mgr)
}
