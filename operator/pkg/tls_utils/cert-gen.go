package tls_utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	fn "github.com/kloudlite/operator/pkg/functions"
)

type GenTLSCertArgs struct {
	// DNSNames is SANs for which certs will be generated
	DNSNames []string

	NotBefore *time.Time
	NotAfter  *time.Time

	CertificateLabel string
}

func GenTLSCert(args GenTLSCertArgs) (caBundle []byte, tlsCert []byte, tlsKey []byte, err error) {
	// Generate a private key for the CA

	if len(args.DNSNames) == 0 {
		return nil, nil, nil, fmt.Errorf("at least 1 SAN must be provided")
	}

	if args.NotBefore == nil {
		args.NotBefore = fn.New(time.Now())
	}

	if args.NotAfter == nil {
		args.NotAfter = fn.New(time.Now().Add(100 * 365 * 24 * time.Hour)) // 100 years
	}

	if args.CertificateLabel == "" {
		args.CertificateLabel = "My Certificate"
	}

	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create a template for the CA certificate
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Kloudlite CA"},
		},
		NotBefore:             *args.NotBefore,
		NotAfter:              args.NotAfter.Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create the CA certificate
	caCertBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPriv.PublicKey, caPriv)
	if err != nil {
		return nil, nil, nil, err
	}

	// Encode the CA certificate to PEM
	caCertPEM := new(bytes.Buffer)
	err = pem.Encode(caCertPEM, &pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes})
	if err != nil {
		// return nil, nil, nil, err
	}

	// Encode the CA private key to PEM
	caKeyPEM := new(bytes.Buffer)
	caPrivBytes, err := x509.MarshalECPrivateKey(caPriv)
	if err != nil {
		return nil, nil, nil, err
	}
	err = pem.Encode(caKeyPEM, &pem.Block{Type: "EC PRIVATE KEY", Bytes: caPrivBytes})
	if err != nil {
		return nil, nil, nil, err
	}

	// Generate a private key for the server
	serverPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create a template for the server certificate
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{args.CertificateLabel},
		},
		NotBefore:   *args.NotBefore,
		NotAfter:    *args.NotAfter,
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    args.DNSNames,
	}

	caCert, err := x509.ParseCertificate(caCertBytes)
	if err != nil {
		return nil, nil, nil, err
	}

	// Create the server certificate
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, caCert, &serverPriv.PublicKey, caPriv)
	if err != nil {
		return nil, nil, nil, err
	}

	// Encode the server certificate to PEM
	serverCertPEM := new(bytes.Buffer)
	err = pem.Encode(serverCertPEM, &pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes})
	if err != nil {
		return nil, nil, nil, err
	}

	// Encode the server private key to PEM
	serverKeyPEM := new(bytes.Buffer)
	serverPrivBytes, err := x509.MarshalECPrivateKey(serverPriv)
	if err != nil {
		return nil, nil, nil, err
	}
	err = pem.Encode(serverKeyPEM, &pem.Block{Type: "EC PRIVATE KEY", Bytes: serverPrivBytes})
	if err != nil {
		return nil, nil, nil, err
	}

	return caCertPEM.Bytes(), serverCertPEM.Bytes(), serverKeyPEM.Bytes(), nil
}
