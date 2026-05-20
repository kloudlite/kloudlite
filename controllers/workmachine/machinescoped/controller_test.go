package machinescoped

import (
	"context"
	"testing"

	workmachinev1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestShouldReconcileWorkMachineUsesNodeWorkMachineLabel(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := workmachinev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	r := &MachineScopedReconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "node-a",
				Labels: map[string]string{"kloudlite.io/workmachine": "wm-a"},
			},
		}).Build(),
		env: Env{NodeName: "node-a"},
	}

	if !r.shouldReconcileWorkMachine(context.Background(), &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm-a"}}) {
		t.Fatal("expected controller to reconcile workmachine selected by node label")
	}
	if r.shouldReconcileWorkMachine(context.Background(), &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm-b"}}) {
		t.Fatal("expected controller to skip workmachine not selected by node label")
	}
}

func TestShouldReconcileWorkMachineKeepsExplicitWorkMachineNameOverride(t *testing.T) {
	r := &MachineScopedReconciler{env: Env{WorkMachineName: "wm-a", NodeName: "node-a"}}

	if !r.shouldReconcileWorkMachine(context.Background(), &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm-a"}}) {
		t.Fatal("expected explicit workmachine name to match")
	}
	if r.shouldReconcileWorkMachine(context.Background(), &workmachinev1.WorkMachine{ObjectMeta: metav1.ObjectMeta{Name: "wm-b"}}) {
		t.Fatal("expected explicit workmachine name to filter other workmachines")
	}
}
