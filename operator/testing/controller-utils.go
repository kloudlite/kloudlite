package testing

import (
	"fmt"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

func Promise(testFn func(g Gomega), timeout ...string) {
	t := 10 * time.Second
	if len(timeout) > 0 {
		t2, err := time.ParseDuration(timeout[0])
		Expect(err).NotTo(HaveOccurred())
		t = t2
	}
	Eventually(func(g Gomega) {
		testFn(g)
	}).WithPolling(100 * time.Millisecond).WithTimeout(t).Should(Succeed())
}

func CreateResource(res client.Object) {
	err := Suite.K8sClient.Create(Suite.Context, res)
	if err != nil {
		if apiErrors.IsAlreadyExists(err) {
			return
		}
	}
	Expect(err).NotTo(HaveOccurred())
}

func DeleteResource(res client.Object) {
	Expect(Suite.K8sClient.Delete(Suite.Context, res)).NotTo(HaveOccurred())
}

func Reconcile(reconciler reconcile.Reconciler, nn types.NamespacedName) {
	_, err := reconciler.Reconcile(Suite.Context, reconcile.Request{NamespacedName: nn})

	Expect(err).ToNot(HaveOccurred(), func() string {
		var t interface{ StackTrace() errors.StackTrace }
		if errors.As(err, &t) {
			return fmt.Sprintf("[partial] error trace:%+v\n", t.StackTrace()[:1])
		}
		return ""
	})
}

func ReconcileForObject(reconciler reconcile.Reconciler, obj client.Object) {
	_, err := reconciler.Reconcile(Suite.Context, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(obj)})

	Expect(err).ToNot(HaveOccurred(), func() string {
		var t interface{ StackTrace() errors.StackTrace }
		if errors.As(err, &t) {
			return fmt.Sprintf("[partial] error trace:%+v\n", t.StackTrace()[:1])
		}
		return ""
	})
}
