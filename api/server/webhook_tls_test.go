package server

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"testing"

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
