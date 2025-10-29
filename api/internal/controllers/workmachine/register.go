package workmachine

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/kubectl"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Register(mgr ctrl.Manager) error {
	var env struct {
		K3sURL        string `envconfig:"K3S_URL"`
		K3sAgentToken string `envconfig:"K3S_AGENT_TOKEN"`
	}

	if err := envconfig.Process("", &env); err != nil {
		return err
	}

	// mgr.AddToSchemes(clustersv1.AddToScheme)
	// mgr.RegisterControllers(
	// 	&nodepool_controller.Reconciler{Env: ev, Name: "nodepool", YAMLClient: mgr.Operator().KubeYAMLClient()},
	// 	&node_controller.Reconciler{Env: ev, Name: "nodepool:node", YAMLClient: mgr.Operator().KubeYAMLClient()},
	// )

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
