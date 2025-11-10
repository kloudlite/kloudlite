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
	"github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type CertificateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *CertificateReconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	req, err := reconciler.NewRequest(ctx, r.Client, request.NamespacedName, &v1.Certificate{})
	if err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	return reconciler.ReconcileSteps(req, []reconciler.Step[*v1.Certificate]{
		{
			Name:     "setup CA and TLS Certificate",
			Title:    "Setup Certificate Authority and TLS Certificate",
			OnCreate: r.createCert,
			OnDelete: r.deleteCert,
		},
	})
}

func (r *CertificateReconciler) createCert(check *reconciler.Check[*v1.Certificate], obj *v1.Certificate) reconciler.StepResult {
	caSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-tls", obj.Name),
			Namespace: obj.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, caSecret, func() error {
		if caSecret.Labels == nil {
			caSecret.Labels = make(map[string]string)
		}
		caSecret.Labels["kloudlite.io/certificate-authority"] = obj.Name
		caSecret.Labels["kloudlite.io/certificate-type"] = "ca"

		// Set owner reference for cascade deletion
		if err := controllerutil.SetControllerReference(obj, caSecret, r.Scheme); err != nil {
			return errors.Wrap("failed to set owner reference", err)
		}

		// Use TLS secret type for Ingress compatibility
		if caSecret.Data == nil {
			caBundle, tlsCert, tlsKey, err := r.genTLSCert(check.Context(), obj)
			if err != nil {
				return err
			}

			caSecret.Data = make(map[string][]byte)
			// Standard TLS secret format for Ingress
			caSecret.Data["tls.crt"] = tlsCert
			caSecret.Data["tls.key"] = tlsKey
			// Also include CA bundle for client trust
			caSecret.Data["ca.crt"] = caBundle
		}

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update CA secret: %w", err))
	}

	return check.Passed()
}

func (r *CertificateReconciler) genTLSCert(ctx context.Context, obj *v1.Certificate) (caBundle, tlsCert, tlsKey []byte, err error) {
	ca := &v1.CertificateAuthority{}
	if err := r.Get(ctx, fn.NN("", obj.Spec.CA), ca); err != nil {
		return nil, nil, nil, err
	}

	caBundleSecret := &corev1.Secret{}
	if err := r.Get(ctx, fn.NN(ca.Status.SecretRef.Namespace, ca.Status.SecretRef.Name), caBundleSecret); err != nil {
		return nil, nil, nil, err
	}

	caBundle, ok := caBundleSecret.Data[ca.Status.SecretRef.CaBundleKey]
	if !ok {
		return nil, nil, nil, fmt.Errorf("CABundle Secret Ref is invalid")
	}

	caPrivateKey, ok := caBundleSecret.Data[ca.Status.SecretRef.CaPrivateKey]
	if !ok {
		return nil, nil, nil, fmt.Errorf("CABundle Secret Ref is invalid")
	}

	// Parse the CA certificate
	caCertBlock, _ := pem.Decode(caBundle)
	if caCertBlock == nil {
		return nil, nil, nil, errors.Wrap("failed to decode CA certificate PEM")
	}

	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, nil, nil, errors.Wrap("failed to parse CA certificate", err)
	}

	// Parse the CA private key
	caKeyBlock, _ := pem.Decode(caPrivateKey)
	if caKeyBlock == nil {
		return nil, nil, nil, fmt.Errorf("failed to decode CA private key PEM")
	}

	caPriv, err := x509.ParseECPrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse CA private key: %w", err)
	}

	// Generate a private key for the server
	serverPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate server private key: %w", err)
	}

	// Create a template for the server certificate
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()), // Use timestamp for unique serial
		Subject: pkix.Name{
			Organization: []string{fmt.Sprintf("Kloudlite Local TLS Cert for %s", obj.Name)},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour * 100 * 365 * 24),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    ca.Spec.SANs,
	}

	// Create the server certificate signed by the CA
	serverCertBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, caCert, &serverPriv.PublicKey, caPriv)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	// Encode the server certificate to PEM
	serverCertPEM := new(bytes.Buffer)
	err = pem.Encode(serverCertPEM, &pem.Block{Type: "CERTIFICATE", Bytes: serverCertBytes})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to encode server certificate: %w", err)
	}

	// Encode the server private key to PEM
	serverKeyPEM := new(bytes.Buffer)
	serverPrivBytes, err := x509.MarshalECPrivateKey(serverPriv)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal server private key: %w", err)
	}
	err = pem.Encode(serverKeyPEM, &pem.Block{Type: "EC PRIVATE KEY", Bytes: serverPrivBytes})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to encode server private key: %w", err)
	}

	return caBundle, serverCertPEM.Bytes(), serverKeyPEM.Bytes(), nil
}

func (r *CertificateReconciler) deleteCert(check *reconciler.Check[*v1.Certificate], obj *v1.Certificate) reconciler.StepResult {
	return check.Passed()
}

func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	builder := ctrl.NewControllerManagedBy(mgr).For(&v1.Certificate{}).Named("certificate")
	builder.Owns(&corev1.Secret{})
	builder.WithEventFilter(reconciler.ReconcileFilter(mgr.GetEventRecorderFor("certificate")))
	return builder.Complete(r)
}
