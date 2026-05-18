package services

import (
	"context"
	"testing"

	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWebhookInstallerLeavesValidExistingCABundleUntouched(t *testing.T) {
	caBundle := []byte("current-ca")
	existing := &admissionv1.ValidatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "kloudlite-validating-webhook",
			Annotations: map[string]string{"keep": "me"},
		},
		Webhooks: []admissionv1.ValidatingWebhook{{
			Name: "users.kloudlite.io",
			ClientConfig: admissionv1.WebhookClientConfig{
				CABundle: caBundle,
			},
		}},
	}
	kube := newWebhookInstallerFakeClient(t, existing)

	installer := NewWebhookInstaller(kube, zap.NewNop(), caBundle)
	if err := installer.createOrUpdateValidatingWebhook(context.Background(), &admissionv1.ValidatingWebhookConfiguration{ObjectMeta: metav1.ObjectMeta{Name: "kloudlite-validating-webhook"}}); err != nil {
		t.Fatalf("create or update validating webhook: %v", err)
	}

	actual := &admissionv1.ValidatingWebhookConfiguration{}
	if err := kube.Get(context.Background(), client.ObjectKey{Name: "kloudlite-validating-webhook"}, actual); err != nil {
		t.Fatalf("get validating webhook: %v", err)
	}
	if actual.Annotations["keep"] != "me" {
		t.Fatal("expected existing valid webhook configuration to be left untouched")
	}
}

func newWebhookInstallerFakeClient(t *testing.T, objects ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}
