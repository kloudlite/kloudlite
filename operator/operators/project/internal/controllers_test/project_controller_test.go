package controllers_test

import (
	"context"
	"fmt"
	"time"

	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	fn "github.com/kloudlite/operator/pkg/functions"
	. "github.com/kloudlite/operator/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var testProjectName = "kl-test-ginkgo-project"
var testAccountRef = "ginkgo-test"

var yamlProject = fmt.Sprintf(`
apiVersion: crds.kloudlite.io/v1
kind: Project
metadata:
  name: %s
  labels:
    sdkfal: asdf
spec:
  accountRef: %s
`, testProjectName, testAccountRef)

var _ = Describe("Project Controller", func() {
	Context("Initialize a Project CR", func() {
		It("Should Succeed", func() {
			ctx, cancelFunc := context.WithTimeout(context.TODO(), 2*time.Second)
			defer cancelFunc()
			err := K8sYamlClient.ApplyYAML(ctx, []byte(yamlProject))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	Context("Controller says", func() {
		It("Namespace created", func() {
			Eventually(func() error {
				ns := &corev1.Namespace{}
				return K8sClient.Get(context.TODO(), fn.NN("", testProjectName), ns)
			}).WithPolling(1 * time.Second).WithTimeout(3 * time.Second).ShouldNot(HaveOccurred())
		})

		It("Service Account created, and has one image pull secrets", func() {
			Eventually(func() error {
				svcAcc := &corev1.ServiceAccount{}
				if err := K8sClient.Get(context.TODO(), fn.NN(testProjectName, "kloudlite-svc-account"), svcAcc); err != nil {
					return err
				}
				if len(svcAcc.ImagePullSecrets) == 0 {
					return fmt.Errorf("at least default image pull secret needs to be there")
				}
				return nil
			}).WithPolling(1 * time.Second).WithTimeout(3 * time.Second).ShouldNot(HaveOccurred())
		})

		It("Harbor Project has been created", func() {
			Eventually(func() error {
				hp := &artifactsv1.HarborProject{}
				return K8sClient.Get(context.TODO(), fn.NN(testProjectName, testAccountRef), hp)
			}).WithPolling(1 * time.Second).WithTimeout(3 * time.Second).ShouldNot(HaveOccurred())
		})

		It("Harbor User Account has been created", func() {
			Eventually(func() error {
				huserAcc := &artifactsv1.HarborUserAccount{}
				return K8sClient.Get(context.TODO(), fn.NN(testProjectName, "kloudlite-docker-registry"), huserAcc)
			}).WithPolling(1 * time.Second).WithTimeout(3 * time.Second).ShouldNot(HaveOccurred())
		})
	})

	Context("Deleting project CR", func() {
		It("should succeed", func() {
			err := K8sYamlClient.DeleteYAML(context.TODO(), []byte(yamlProject))
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
