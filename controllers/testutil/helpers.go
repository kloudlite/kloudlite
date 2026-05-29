package testutil

import (
	environmentsv1 "github.com/kloudlite/kloudlite/types/environment/v1"
	packagesv1 "github.com/kloudlite/kloudlite/types/packages/v1"
	snapshotsv1 "github.com/kloudlite/kloudlite/types/snapshot/v1"
	machinesv1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	workspacesv1 "github.com/kloudlite/kloudlite/types/workspace/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// NewTestScheme creates a new runtime scheme with all required types
func NewTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = environmentsv1.AddToScheme(scheme)
	_ = packagesv1.AddToScheme(scheme)
	_ = snapshotsv1.AddToScheme(scheme)
	_ = machinesv1.AddToScheme(scheme)
	_ = workspacesv1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	_ = rbacv1.AddToScheme(scheme)
	return scheme
}

// NewFakeClient creates a fake Kubernetes client for testing
func NewFakeClient(scheme *runtime.Scheme, objs ...client.Object) *fake.ClientBuilder {
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...)
}

// Int32Ptr returns a pointer to an int32 value
func Int32Ptr(i int32) *int32 {
	return &i
}
