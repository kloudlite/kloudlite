package certs

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// WebhookCertificates holds the generated certificates for webhooks
type WebhookCertificates struct {
	CACert     []byte // PEM encoded CA certificate
	ServerCert []byte // PEM encoded server certificate
	ServerKey  []byte // PEM encoded server private key
	CABundle   string // Base64 encoded CA certificate for webhook config
}

// GenerateWebhookCertificates generates self-signed certificates for webhook server
func GenerateWebhookCertificates(serviceName, namespace string) (*WebhookCertificates, error) {
	// Generate CA private key
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA key: %w", err)
	}

	// Create CA certificate template
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Kloudlite Webhook CA",
			Organization: []string{"Kloudlite"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // Valid for 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Self-sign the CA certificate
	caCertBytes, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	// PEM encode CA certificate
	caCertPEM := new(bytes.Buffer)
	if err := pem.Encode(caCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertBytes,
	}); err != nil {
		return nil, fmt.Errorf("failed to encode CA certificate: %w", err)
	}

	// Generate server private key
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server key: %w", err)
	}

	// DNS names for the webhook service
	dnsNames := []string{
		serviceName,
		fmt.Sprintf("%s.%s", serviceName, namespace),
		fmt.Sprintf("%s.%s.svc", serviceName, namespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace),
	}

	// Create server certificate template
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName:   fmt.Sprintf("%s.%s.svc", serviceName, namespace),
			Organization: []string{"Kloudlite"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0), // Valid for 10 years
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    dnsNames,
	}

	// Sign server certificate with CA
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	// PEM encode server certificate
	serverCertPEM := new(bytes.Buffer)
	if err := pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	}); err != nil {
		return nil, fmt.Errorf("failed to encode server certificate: %w", err)
	}

	// PEM encode server private key
	serverKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(serverKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverKey),
	}); err != nil {
		return nil, fmt.Errorf("failed to encode server key: %w", err)
	}

	// Base64 encode CA certificate for webhook configuration
	caBundle := base64.StdEncoding.EncodeToString(caCertPEM.Bytes())

	return &WebhookCertificates{
		CACert:     caCertPEM.Bytes(),
		ServerCert: serverCertPEM.Bytes(),
		ServerKey:  serverKeyPEM.Bytes(),
		CABundle:   caBundle,
	}, nil
}
