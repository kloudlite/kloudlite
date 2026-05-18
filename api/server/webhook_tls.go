package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const webhookTLSSecretName = "platform-api-tls"
const webhookTLSRotationAnnotation = "kloudlite.io/platform-api-tls-rotation-id"
const frontendTLSRestartAnnotation = "kloudlite.io/platform-api-tls-restarted-at"

const (
	webhookTLSSecretCreated  = "created"
	webhookTLSSecretReused   = "reused_existing"
	webhookTLSSecretReplaced = "replaced_invalid"
)

type webhookTLSOptions struct {
	Namespace   string
	ServiceName string
}

type webhookTLSBundle struct {
	CertPEM     []byte
	KeyPEM      []byte
	Certificate tls.Certificate
	Source      string
}

func ensureWebhookTLSSecret(ctx context.Context, kube client.Client, opts webhookTLSOptions) (*webhookTLSBundle, error) {
	secret := &corev1.Secret{}
	key := client.ObjectKey{Namespace: opts.Namespace, Name: webhookTLSSecretName}
	if err := kube.Get(ctx, key, secret); err == nil {
		bundle, bundleErr := webhookTLSBundleFromSecret(secret, opts)
		if bundleErr == nil {
			if rotationID := secret.Annotations[webhookTLSRotationAnnotation]; rotationID != "" {
				if err := restartFrontendForWebhookTLSRotation(ctx, kube, opts.Namespace, rotationID); err != nil {
					return nil, err
				}
			}
			bundle.Source = webhookTLSSecretReused
			return bundle, nil
		}

		bundle, err = generateWebhookTLSBundle(opts)
		if err != nil {
			return nil, err
		}
		secret.Type = corev1.SecretTypeTLS
		if secret.Annotations == nil {
			secret.Annotations = map[string]string{}
		}
		rotationID := time.Now().UTC().Format(time.RFC3339Nano)
		secret.Annotations[webhookTLSRotationAnnotation] = rotationID
		secret.Data = webhookTLSSecretData(bundle)
		if err := kube.Update(ctx, secret); err != nil {
			return nil, fmt.Errorf("update webhook TLS secret: %w", err)
		}
		if err := restartFrontendForWebhookTLSRotation(ctx, kube, opts.Namespace, rotationID); err != nil {
			return nil, err
		}
		bundle.Source = webhookTLSSecretReplaced
		return bundle, nil
	} else if !apierrors.IsNotFound(err) {
		return nil, fmt.Errorf("get webhook TLS secret: %w", err)
	}

	bundle, err := generateWebhookTLSBundle(opts)
	if err != nil {
		return nil, err
	}
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      webhookTLSSecretName,
			Namespace: opts.Namespace,
			Annotations: map[string]string{
				webhookTLSRotationAnnotation: time.Now().UTC().Format(time.RFC3339Nano),
			},
		},
		Type: corev1.SecretTypeTLS,
		Data: webhookTLSSecretData(bundle),
	}
	if err := kube.Create(ctx, secret); err != nil {
		return nil, fmt.Errorf("create webhook TLS secret: %w", err)
	}
	bundle.Source = webhookTLSSecretCreated
	return bundle, nil
}

func restartFrontendForWebhookTLSRotation(ctx context.Context, kube client.Client, namespace, rotationID string) error {
	deployment := &appsv1.Deployment{}
	key := client.ObjectKey{Namespace: namespace, Name: "frontend"}
	if err := kube.Get(ctx, key, deployment); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("get frontend deployment for webhook TLS rotation: %w", err)
	}
	patched := deployment.DeepCopy()
	if patched.Spec.Template.Annotations == nil {
		patched.Spec.Template.Annotations = map[string]string{}
	}
	if patched.Spec.Template.Annotations[frontendTLSRestartAnnotation] == rotationID {
		return nil
	}
	patched.Spec.Template.Annotations[frontendTLSRestartAnnotation] = rotationID
	if err := kube.Patch(ctx, patched, client.MergeFrom(deployment)); err != nil {
		return fmt.Errorf("restart frontend deployment for webhook TLS rotation: %w", err)
	}
	return nil
}

func webhookTLSBundleFromSecret(secret *corev1.Secret, opts webhookTLSOptions) (*webhookTLSBundle, error) {
	certPEM := secret.Data[corev1.TLSCertKey]
	keyPEM := secret.Data[corev1.TLSPrivateKeyKey]
	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	leaf, err := parseWebhookTLSLeaf(certPEM)
	if err != nil {
		return nil, err
	}
	if time.Now().After(leaf.NotAfter) {
		return nil, fmt.Errorf("webhook TLS certificate expired at %s", leaf.NotAfter.Format(time.RFC3339))
	}
	for _, name := range webhookTLSDNSNames(opts) {
		if err := leaf.VerifyHostname(name); err != nil {
			return nil, fmt.Errorf("webhook TLS certificate missing DNS name %q: %w", name, err)
		}
	}
	return &webhookTLSBundle{CertPEM: certPEM, KeyPEM: keyPEM, Certificate: certificate}, nil
}

func parseWebhookTLSLeaf(certPEM []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, fmt.Errorf("parse webhook TLS certificate: missing PEM block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse webhook TLS certificate: %w", err)
	}
	return cert, nil
}

func webhookTLSSecretData(bundle *webhookTLSBundle) map[string][]byte {
	return map[string][]byte{
		corev1.TLSCertKey:       bundle.CertPEM,
		corev1.TLSPrivateKeyKey: bundle.KeyPEM,
	}
}

func generateWebhookTLSBundle(opts webhookTLSOptions) (*webhookTLSBundle, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate webhook TLS key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("generate webhook TLS serial: %w", err)
	}

	dnsNames := webhookTLSDNSNames(opts)
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("%s.%s.svc", opts.ServiceName, opts.Namespace),
		},
		DNSNames:              dnsNames,
		NotBefore:             time.Now().Add(-time.Minute),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("create webhook TLS certificate: %w", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("load generated webhook TLS key pair: %w", err)
	}
	return &webhookTLSBundle{CertPEM: certPEM, KeyPEM: keyPEM, Certificate: certificate}, nil
}

func webhookTLSDNSNames(opts webhookTLSOptions) []string {
	return []string{
		opts.ServiceName,
		fmt.Sprintf("%s.%s", opts.ServiceName, opts.Namespace),
		fmt.Sprintf("%s.%s.svc", opts.ServiceName, opts.Namespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", opts.ServiceName, opts.Namespace),
	}
}
