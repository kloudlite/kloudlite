package certs

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

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/certs/v1"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CAReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *CAReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
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

func (r *CAReconciler) createCAandCert(check *reconciler.Check[*v1.CertificateAuthority], obj *v1.CertificateAuthority) reconciler.StepResult {
	caCertSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      obj.Name,
			Namespace: "kloudlite-ingress",
		},
	}

	const caBundleKey = "ca.crt"
	const caPrivateKey = "ca.key"

	obj.Status.SecretRef = v1.SecretKeyRef{
		Name:         caCertSecret.Name,
		Namespace:    caCertSecret.Namespace,
		CaBundleKey:  caBundleKey,
		CaPrivateKey: caPrivateKey,
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, caCertSecret, func() error {
		if caCertSecret.Labels == nil {
			caCertSecret.Labels = make(map[string]string)
		}
		caCertSecret.Labels["kloudlite.io/certificate-authority"] = obj.Name
		caCertSecret.Labels["kloudlite.io/certificate-type"] = "ca"

		// Set owner reference for cascade deletion
		if err := controllerutil.SetControllerReference(obj, caCertSecret, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference: %w", err)
		}

		// Use TLS secret type for Ingress compatibility
		if caCertSecret.Data == nil {
			caCertSecret.Data = make(map[string][]byte)

			caBundle, caKey, err := r.generateCABundle(check.Context(), obj)
			if err != nil {
				return err
			}

			caCertSecret.Data[caBundleKey] = caBundle
			caCertSecret.Data[caPrivateKey] = caKey
		}

		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

func (r *CAReconciler) deleteCAandCert(check *reconciler.Check[*v1.CertificateAuthority], obj *v1.CertificateAuthority) reconciler.StepResult {
	return check.Passed()
}

func (r *CAReconciler) generateCABundle(ctx context.Context, obj *v1.CertificateAuthority) (caBundle, caPrivateKey []byte, err error) {
	// Generate a private key for the CA
	caPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Create a template for the CA certificate
	caTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{"Kloudlite CA"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(100 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		Issuer: pkix.Name{
			Country:       []string{"IN"},
			Organization:  []string{"Kloudlite"},
			Locality:      []string{},
			Province:      []string{"Karnataka"},
			StreetAddress: []string{"415, VRR Fortuna", "Janatha Colony, Bangalore"},
			PostalCode:    []string{"560035"},
			CommonName:    "kloudlite",
		},
	}

	// Create the CA certificate
	caCertBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPriv.PublicKey, caPriv)
	if err != nil {
		return nil, nil, err
	}

	// Encode the CA certificate to PEM
	caCertPEM := new(bytes.Buffer)
	err = pem.Encode(caCertPEM, &pem.Block{Type: "CERTIFICATE", Bytes: caCertBytes})
	if err != nil {
		return nil, nil, err
	}

	caPrivateKeyPEM := new(bytes.Buffer)
	caPrivateKeyBytes, err := x509.MarshalECPrivateKey(caPriv)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal CA private key: %w", err)
	}
	err = pem.Encode(caPrivateKeyPEM, &pem.Block{Type: "EC PRIVATE KEY", Bytes: caPrivateKeyBytes})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode CA private key: %w", err)
	}

	return caCertPEM.Bytes(), caPrivateKeyPEM.Bytes(), nil
}

func (r *CAReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.CertificateAuthority{}).Named("certificate-authority")
	builder.Owns(&corev1.Secret{})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor("certificate-authority")))
	return builder.Complete(r)
}
