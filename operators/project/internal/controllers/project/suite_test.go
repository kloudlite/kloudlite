package project

import (
	"testing"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/project/internal/env"
	"github.com/kloudlite/operator/pkg/logging"

	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var reconciler *Reconciler

var _ = BeforeSuite(func() {
	SetupKubernetes(AddToSchemes(crdsv1.AddToScheme), DefaultEnvTest)
	reconciler = &Reconciler{
		Client:     nil,
		Scheme:     Suite.Scheme,
		logger:     logging.NewOrDie(&logging.Options{Name: "project", Dev: true}),
		Name:       "project",
		yamlClient: Suite.K8sYamlClient,
		Env: &env.Env{
			ReconcilePeriod:         30 * time.Second,
			MaxConcurrentReconciles: 10,
			DockerSecretName:        "harbor-admin-creds",
			AdminRoleName:           "harbor-admin-role",
			SvcAccountName:          "kloudlite-svc-account",
		},
	}
})
