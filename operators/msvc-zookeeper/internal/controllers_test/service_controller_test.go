package controllers_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	zookeeperMsvcv1 "operators.kloudlite.io/apis/zookeeper.msvc/v1"
	. "operators.kloudlite.io/test-lib"
)

var namespace = "ginkgo-test-1"
var zookeeperName = "sample-zookeeper"

var _ = Describe("Zookeeper Controller", func() {
	Context("Initializes Zookeeper CR", func() {
		It("Should Succeed", func() {
			ctx, cancelFunc := context.WithTimeout(context.TODO(), 2*time.Second)
			defer cancelFunc()
			err := K8sClient.Create(ctx, &zookeeperMsvcv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      zookeeperName,
					Namespace: namespace,
					Labels: map[string]string{
						"kloudlite.io/created-in-test": "true",
					},
				},
				Spec: zookeeperMsvcv1.ServiceSpec{
					Region:       "",
					ReplicaCount: 1,
					Resources: ct.Resources{
						Cpu:    ct.CpuT{Min: "400m", Max: "500m"},
						Memory: "800Mi",
						Storage: &ct.Storage{
							Size:         "1Gi",
							StorageClass: "local-path",
						},
					},
				},
			})
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})
