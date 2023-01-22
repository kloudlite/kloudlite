package project

import (
	v1 "github.com/kloudlite/operator/apis/crds/v1"
	. "github.com/onsi/ginkgo/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

func newProject(name, namespace string) v1.Project {
	return v1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.ProjectSpec{
			AccountRef:  "kl-core",
			DisplayName: rand.String(10),
		},
	}
}

var _ = Describe("project controller says,", func() {
	It("creates blueprint namespace", func() {
		Fail("empty test")
	})

	It("creates a default environment referenced to blueprint, if not exisds", func() {
		Fail("empty test")
	})

	It("create harbor project and user account", func() {
		Fail("empty test")
	})

	It("project is marked as Ready, if everything succeeds", func() {
		Fail("empty test")
	})
})
