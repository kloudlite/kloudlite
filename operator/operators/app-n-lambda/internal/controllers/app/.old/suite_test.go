/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package _old

import (
	"context"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"github.com/kloudlite/operator/pkg/kubectl"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sClient     client.Client
	testEnv       *envtest.Environment
	k8sYamlClient *kubectl.YAMLClient
)

func GetObject[T client.Object](ctx context.Context, nn types.NamespacedName, obj T) (T, error) {
	if err := k8sClient.Get(ctx, nn, obj); err != nil {
		return *new(T), err
	}
	return obj, nil
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

const YamlTestNs = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: ginkgo-test-1
`

var _ = BeforeSuite(
	func() {
		Expect(os.Setenv("USE_EXISTING_CLUSTER", "true")).To(Succeed())
		logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

		By("bootstrapping test environment")
		testEnv = &envtest.Environment{
			Config: &rest.Config{Host: "localhost:8080"},
			// CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
			// ErrorIfCRDPathMissing: true,
		}

		cfg, err := testEnv.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).NotTo(BeNil())

		err = crdsv1.AddToScheme(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		// +kubebuilder:scaffold:scheme

		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient).NotTo(BeNil())

		k8sYamlClient, err = kubectl.NewYAMLClient(cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sYamlClient).NotTo(BeNil())

		err = k8sYamlClient.ApplyYAML(context.TODO(), []byte(YamlTestNs))
		Expect(err).NotTo(HaveOccurred())
	},
)

var _ = AfterSuite(
	func() {
		By("tearing down the test environment")
		err := testEnv.Stop()
		Expect(err).NotTo(HaveOccurred())
	},
)
