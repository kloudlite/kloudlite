package services

import (
	"context"

	cav1 "github.com/kloudlite/kloudlite/api/internal/controllers/certs/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	"go.uber.org/zap"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kloudliteCAName          = "kloudlite-ca"
	kloudliteCertificateName = "kloudlite-wildcard-cert"
)

// CAInitializer handles initialization of the Kloudlite CA and wildcard certificate
type CAInitializer struct {
	k8sClient client.Client
	logger    *zap.Logger
}

// NewCAInitializer creates a new CA initializer
func NewCAInitializer(k8sClient client.Client, logger *zap.Logger) *CAInitializer {
	return &CAInitializer{
		k8sClient: k8sClient,
		logger:    logger.Named("ca-initializer"),
	}
}

// ensureCA creates or updates the Kloudlite CertificateAuthority
func (c *CAInitializer) ensureCA(ctx context.Context, subdomain string) error {
	ca := &cav1.CertificateAuthority{
		ObjectMeta: metav1.ObjectMeta{
			Name: kloudliteCAName,
		},
	}

	if err := c.k8sClient.Get(ctx, client.ObjectKeyFromObject(ca), ca); err != nil {
		if !apiErrors.IsNotFound(err) {
			c.logger.Error("[ensureCA] failed", zap.Error(err))
			return err
		}
		ca.Spec.SANs = []string{subdomain}
		if err := c.k8sClient.Create(ctx, ca); err != nil {
			return errors.Wrap("failed to create Certificate Authority", err)
		}
		c.logger.Info("Created CertificateAuthority", zap.String("name", kloudliteCAName), zap.Strings("sans", ca.Spec.SANs))
	}

	c.logger.Info("Ensured CertificateAuthority", zap.String("name", kloudliteCAName), zap.Strings("sans", ca.Spec.SANs))
	return nil
}

// ensureWildcardCertificate creates or updates the wildcard Certificate
func (c *CAInitializer) ensureWildcardCertificate(ctx context.Context) error {
	certificate := &cav1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kloudliteCertificateName,
			Namespace: "kloudlite-ingress",
		},
	}

	if err := c.k8sClient.Get(ctx, client.ObjectKeyFromObject(certificate), certificate); err != nil {
		if !apiErrors.IsNotFound(err) {
			c.logger.Error("[ensureWildcardCertificate] failed", zap.Error(err))
			return err
		}

		certificate.Spec.CA = kloudliteCAName
		if err := c.k8sClient.Create(ctx, certificate); err != nil {
			return errors.Wrap("failed to create Certificate", err)
		}
		c.logger.Info("Created wildcard Certificate",
			zap.String("name", kloudliteCertificateName),
			zap.String("namespace", "kloudlite-ingress"),
		)
	}

	c.logger.Info("Ensured wildcard Certificate",
		zap.String("name", kloudliteCertificateName),
		zap.String("namespace", "kloudlite-ingress"),
	)

	return nil
}
