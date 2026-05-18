package server

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestEnsureWebhookTLSSecretCreatesStableSecretWhenMissing(t *testing.T) {
	kube := newServerFakeClient(t)

	bundle, err := ensureWebhookTLSSecret(context.Background(), kube, webhookTLSOptions{Namespace: "kloudlite", ServiceName: "api-server"})
	if err != nil {
		t.Fatalf("ensure webhook TLS secret: %v", err)
	}

	secret := &corev1.Secret{}
	if err := kube.Get(context.Background(), client.ObjectKey{Namespace: "kloudlite", Name: webhookTLSSecretName}, secret); err != nil {
		t.Fatalf("expected TLS secret to be created: %v", err)
	}
	if secret.Type != corev1.SecretTypeTLS {
		t.Fatalf("expected TLS secret type, got %q", secret.Type)
	}
	if string(secret.Data[corev1.TLSCertKey]) != string(bundle.CertPEM) {
		t.Fatal("expected returned cert to match secret tls.crt")
	}
	if string(secret.Data[corev1.TLSPrivateKeyKey]) != string(bundle.KeyPEM) {
		t.Fatal("expected returned key to match secret tls.key")
	}
	if len(bundle.Certificate.Certificate) == 0 {
		t.Fatal("expected in-memory TLS certificate")
	}
	if bundle.Source != webhookTLSSecretCreated {
		t.Fatalf("expected source %q, got %q", webhookTLSSecretCreated, bundle.Source)
	}

	parsed := parseFirstCertificate(t, bundle.CertPEM)
	assertDNSName(t, parsed.DNSNames, "api-server.kloudlite.svc.cluster.local")

	second, err := ensureWebhookTLSSecret(context.Background(), kube, webhookTLSOptions{Namespace: "kloudlite", ServiceName: "api-server"})
	if err != nil {
		t.Fatalf("ensure existing webhook TLS secret: %v", err)
	}
	if string(second.CertPEM) != string(bundle.CertPEM) {
		t.Fatal("expected existing valid secret to be reused")
	}
	if second.Source != webhookTLSSecretReused {
		t.Fatalf("expected source %q, got %q", webhookTLSSecretReused, second.Source)
	}
}

func TestEnsureWebhookTLSSecretReplacesInvalidSecret(t *testing.T) {
	invalid := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: webhookTLSSecretName, Namespace: "kloudlite"},
		Type:       corev1.SecretTypeTLS,
		Data: map[string][]byte{
			corev1.TLSCertKey:       []byte("not a cert"),
			corev1.TLSPrivateKeyKey: []byte("not a key"),
		},
	}
	kube := newServerFakeClient(t, invalid)

	bundle, err := ensureWebhookTLSSecret(context.Background(), kube, webhookTLSOptions{Namespace: "kloudlite", ServiceName: "api-server"})
	if err != nil {
		t.Fatalf("ensure webhook TLS secret: %v", err)
	}

	secret := &corev1.Secret{}
	if err := kube.Get(context.Background(), client.ObjectKey{Namespace: "kloudlite", Name: webhookTLSSecretName}, secret); err != nil {
		t.Fatalf("expected TLS secret: %v", err)
	}
	if string(secret.Data[corev1.TLSCertKey]) != string(bundle.CertPEM) {
		t.Fatal("expected invalid secret cert to be replaced")
	}
	if string(secret.Data[corev1.TLSCertKey]) == "not a cert" {
		t.Fatal("expected invalid cert data to be removed")
	}
	if bundle.Source != webhookTLSSecretReplaced {
		t.Fatalf("expected source %q, got %q", webhookTLSSecretReplaced, bundle.Source)
	}
}

func TestEnsureWebhookTLSSecretReplacesSecretForWrongServiceDNS(t *testing.T) {
	wrongBundle, err := generateWebhookTLSBundle(webhookTLSOptions{Namespace: "default", ServiceName: "api-server"})
	if err != nil {
		t.Fatalf("generate wrong namespace bundle: %v", err)
	}
	stale := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: webhookTLSSecretName, Namespace: "kloudlite"},
		Type:       corev1.SecretTypeTLS,
		Data:       webhookTLSSecretData(wrongBundle),
	}
	kube := newServerFakeClient(t, stale)

	bundle, err := ensureWebhookTLSSecret(context.Background(), kube, webhookTLSOptions{Namespace: "kloudlite", ServiceName: "api-server"})
	if err != nil {
		t.Fatalf("ensure webhook TLS secret: %v", err)
	}

	if string(bundle.CertPEM) == string(wrongBundle.CertPEM) {
		t.Fatal("expected secret with wrong service DNS to be replaced")
	}
	if bundle.Source != webhookTLSSecretReplaced {
		t.Fatalf("expected source %q, got %q", webhookTLSSecretReplaced, bundle.Source)
	}
	parsed := parseFirstCertificate(t, bundle.CertPEM)
	assertDNSName(t, parsed.DNSNames, "api-server.kloudlite.svc.cluster.local")
}

func TestEnsureWebhookTLSSecretRestartsFrontendWhenReplacingSecret(t *testing.T) {
	wrongBundle, err := generateWebhookTLSBundle(webhookTLSOptions{Namespace: "default", ServiceName: "api-server"})
	if err != nil {
		t.Fatalf("generate wrong namespace bundle: %v", err)
	}
	stale := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: webhookTLSSecretName, Namespace: "kloudlite"},
		Type:       corev1.SecretTypeTLS,
		Data:       webhookTLSSecretData(wrongBundle),
	}
	frontend := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: "kloudlite"},
	}
	kube := newServerFakeClient(t, stale, frontend)

	_, err = ensureWebhookTLSSecret(context.Background(), kube, webhookTLSOptions{Namespace: "kloudlite", ServiceName: "api-server"})
	if err != nil {
		t.Fatalf("ensure webhook TLS secret: %v", err)
	}

	updated := &appsv1.Deployment{}
	if err := kube.Get(context.Background(), client.ObjectKey{Namespace: "kloudlite", Name: "frontend"}, updated); err != nil {
		t.Fatalf("get frontend deployment: %v", err)
	}
	if updated.Spec.Template.Annotations[frontendTLSRestartAnnotation] == "" {
		t.Fatal("expected frontend deployment restart annotation after TLS secret replacement")
	}
}

func TestEnsureWebhookTLSSecretRetriesPendingFrontendRestart(t *testing.T) {
	bundle, err := generateWebhookTLSBundle(webhookTLSOptions{Namespace: "kloudlite", ServiceName: "api-server"})
	if err != nil {
		t.Fatalf("generate bundle: %v", err)
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webhookTLSSecretName,
			Namespace: "kloudlite",
			Annotations: map[string]string{
				webhookTLSRotationAnnotation: "rotation-1",
			},
		},
		Type: corev1.SecretTypeTLS,
		Data: webhookTLSSecretData(bundle),
	}
	frontend := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: "kloudlite"},
	}
	kube := newServerFakeClient(t, secret, frontend)

	reused, err := ensureWebhookTLSSecret(context.Background(), kube, webhookTLSOptions{Namespace: "kloudlite", ServiceName: "api-server"})
	if err != nil {
		t.Fatalf("ensure webhook TLS secret: %v", err)
	}
	if reused.Source != webhookTLSSecretReused {
		t.Fatalf("expected source %q, got %q", webhookTLSSecretReused, reused.Source)
	}

	updated := &appsv1.Deployment{}
	if err := kube.Get(context.Background(), client.ObjectKey{Namespace: "kloudlite", Name: "frontend"}, updated); err != nil {
		t.Fatalf("get frontend deployment: %v", err)
	}
	if updated.Spec.Template.Annotations[frontendTLSRestartAnnotation] != "rotation-1" {
		t.Fatalf("expected frontend restart annotation %q, got %q", "rotation-1", updated.Spec.Template.Annotations[frontendTLSRestartAnnotation])
	}
}

func parseFirstCertificate(t *testing.T, certPEM []byte) *x509.Certificate {
	t.Helper()
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("expected PEM certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	return cert
}

func assertDNSName(t *testing.T, names []string, want string) {
	t.Helper()
	for _, name := range names {
		if name == want {
			return
		}
	}
	t.Fatalf("expected DNS names %v to include %q", names, want)
}
