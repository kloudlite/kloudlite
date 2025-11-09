package certificateauthority

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/certificate-authority/v1"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.CertificateAuthority{})
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.CertificateAuthority]{
		{
			Name:     "setup CA and TLS Certificate",
			Title:    "Setup Certificate Authority and TLS Certificate",
			OnCreate: r.createCAandCert,
			OnDelete: r.deleteCAandCert,
		},
	})
}

func (r *Reconciler) createCAandCert(check *reconciler.Check[*v1.CertificateAuthority], obj *v1.CertificateAuthority) reconciler.StepResult {
	caSecretName := fmt.Sprintf("%s-tls", obj.Name)

	obj.Status.TLSSecretName = caSecretName

	caBundle, tlsCert, tlsKey, err := GenTLSCert(GenTLSCertArgs{
		DNSNames:         obj.Spec.SANs,
		CertificateLabel: "Kloudlite Workmachine Cert",
	})
	if err != nil {
		return check.Failed(fmt.Errorf("failed to generate CA and Certificate: %w", err))
	}

	// Create or update CA secret
	caSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      caSecretName,
			Namespace: obj.Namespace,
		},
		Type: corev1.SecretTypeTLS,
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, caSecret, func() error {
		if caSecret.Labels == nil {
			caSecret.Labels = make(map[string]string)
		}
		caSecret.Labels["kloudlite.io/certificate-authority"] = obj.Name
		caSecret.Labels["kloudlite.io/certificate-type"] = "ca"

		// Set owner reference for cascade deletion
		if err := controllerutil.SetControllerReference(obj, caSecret, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

		// Use TLS secret type for Ingress compatibility
		if caSecret.Data == nil {
			caSecret.Data = make(map[string][]byte)
		}

		// Standard TLS secret format for Ingress
		caSecret.Data["tls.crt"] = tlsCert
		caSecret.Data["tls.key"] = tlsKey
		// Also include CA bundle for client trust
		caSecret.Data["ca.crt"] = caBundle

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update CA secret: %w", err))
	}

	return check.Passed()
}

func (r *Reconciler) deleteCAandCert(check *reconciler.Check[*v1.CertificateAuthority], obj *v1.CertificateAuthority) reconciler.StepResult {
	return check.Passed()
}

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
		now := time.Now()
		args.NotBefore = &now
	}

	if args.NotAfter == nil {
		notAfter := time.Now().Add(100 * 365 * 24 * time.Hour) // 100 years
		args.NotAfter = &notAfter
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
		NotAfter:              *args.NotAfter,
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
		return nil, nil, nil, err
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

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.CertificateAuthority{}).Named("certificate-authority")
	builder.Owns(&corev1.Secret{})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor("certificate-authority")))
	return builder.Complete(r)
}
