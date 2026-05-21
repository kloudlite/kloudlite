package machinescoped

import (
	"github.com/kloudlite/kloudlite/controllers/controllerconfig"
	"github.com/kloudlite/kloudlite/pkg/operator-toolkit/kubectl"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Register(mgr ctrl.Manager, cfg *controllerconfig.ControllerConfig) error {
	yamlClient, err := kubectl.NewYAMLClient(mgr.GetConfig(), kubectl.YAMLClientOpts{})
	if err != nil {
		return err
	}

	return newMachineScopedReconciler(mgr.GetClient(), mgr.GetScheme(), yamlClient, cfg).SetupWithManager(mgr)
}

func newMachineScopedReconciler(client client.Client, scheme *runtime.Scheme, yamlClient kubectl.YAMLClient, cfg *controllerconfig.ControllerConfig) *MachineScopedReconciler {
	return &MachineScopedReconciler{
		Client:     client,
		Scheme:     scheme,
		YAMLClient: yamlClient,
		Cfg:        cfg,
	}
}
