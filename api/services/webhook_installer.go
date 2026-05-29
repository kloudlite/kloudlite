package services

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"

	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed webhook_configs.yaml
var webhookConfigsYAML string

type WebhookInstaller struct {
	k8sClient client.Client
	logger    *zap.Logger
	caBundle  []byte
}

func NewWebhookInstaller(k8sClient client.Client, logger *zap.Logger, caBundle []byte) *WebhookInstaller {
	return &WebhookInstaller{
		k8sClient: k8sClient,
		logger:    logger.Named("webhook-installer"),
		caBundle:  caBundle,
	}
}

// InstallWebhooks installs ValidatingWebhookConfiguration and MutatingWebhookConfiguration
func (wi *WebhookInstaller) InstallWebhooks(ctx context.Context) error {
	wi.logger.Info("ensuring admission webhook configurations trust current platform-api TLS certificate")

	// Split the YAML into separate documents
	documents := strings.Split(webhookConfigsYAML, "\n---\n")

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" || doc == "---" {
			continue
		}

		// Decode the document to determine its type
		decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(doc), 4096)

		var meta metav1.TypeMeta
		if err := decoder.Decode(&meta); err != nil {
			wi.logger.Warn("Failed to decode TypeMeta", zap.Error(err))
			continue
		}

		// Reset decoder to read the full object
		decoder = yaml.NewYAMLOrJSONDecoder(strings.NewReader(doc), 4096)

		switch meta.Kind {
		case "ValidatingWebhookConfiguration":
			var vwc admissionv1.ValidatingWebhookConfiguration
			if err := decoder.Decode(&vwc); err != nil {
				return fmt.Errorf("failed to decode ValidatingWebhookConfiguration: %w", err)
			}

			// Set CA bundle for all webhooks
			for i := range vwc.Webhooks {
				vwc.Webhooks[i].ClientConfig.CABundle = wi.caBundle
			}

			if err := wi.createOrUpdateValidatingWebhook(ctx, &vwc); err != nil {
				return fmt.Errorf("failed to create/update ValidatingWebhookConfiguration: %w", err)
			}

		case "MutatingWebhookConfiguration":
			var mwc admissionv1.MutatingWebhookConfiguration
			if err := decoder.Decode(&mwc); err != nil {
				return fmt.Errorf("failed to decode MutatingWebhookConfiguration: %w", err)
			}

			// Set CA bundle for all webhooks
			for i := range mwc.Webhooks {
				mwc.Webhooks[i].ClientConfig.CABundle = wi.caBundle
			}

			if err := wi.createOrUpdateMutatingWebhook(ctx, &mwc); err != nil {
				return fmt.Errorf("failed to create/update MutatingWebhookConfiguration: %w", err)
			}
		}
	}

	wi.logger.Info("admission webhook configurations are ready")
	return nil
}

func (wi *WebhookInstaller) createOrUpdateValidatingWebhook(ctx context.Context, vwc *admissionv1.ValidatingWebhookConfiguration) error {
	existing := &admissionv1.ValidatingWebhookConfiguration{}
	err := wi.k8sClient.Get(ctx, client.ObjectKey{Name: vwc.Name}, existing)

	if errors.IsNotFound(err) {
		wi.logger.Info("creating validating webhook configuration with current CA bundle", zap.String("name", vwc.Name))
		return wi.k8sClient.Create(ctx, vwc)
	} else if err != nil {
		return err
	}

	if validatingWebhookMatches(existing, vwc, wi.caBundle) {
		wi.logger.Info("validating webhook configuration already matches desired configuration", zap.String("name", vwc.Name))
		return nil
	}
	wi.logger.Info("updating validating webhook configuration to match desired configuration", zap.String("name", vwc.Name))
	vwc.ResourceVersion = existing.ResourceVersion
	return wi.k8sClient.Update(ctx, vwc)
}

func (wi *WebhookInstaller) createOrUpdateMutatingWebhook(ctx context.Context, mwc *admissionv1.MutatingWebhookConfiguration) error {
	existing := &admissionv1.MutatingWebhookConfiguration{}
	err := wi.k8sClient.Get(ctx, client.ObjectKey{Name: mwc.Name}, existing)

	if errors.IsNotFound(err) {
		wi.logger.Info("creating mutating webhook configuration with current CA bundle", zap.String("name", mwc.Name))
		return wi.k8sClient.Create(ctx, mwc)
	} else if err != nil {
		return err
	}

	if mutatingWebhookMatches(existing, mwc, wi.caBundle) {
		wi.logger.Info("mutating webhook configuration already matches desired configuration", zap.String("name", mwc.Name))
		return nil
	}
	wi.logger.Info("updating mutating webhook configuration to match desired configuration", zap.String("name", mwc.Name))
	mwc.ResourceVersion = existing.ResourceVersion
	return wi.k8sClient.Update(ctx, mwc)
}

func validatingWebhookMatches(existing, desired *admissionv1.ValidatingWebhookConfiguration, caBundle []byte) bool {
	return validatingWebhookCABundleValid(existing, caBundle) && equality.Semantic.DeepEqual(existing.Webhooks, desired.Webhooks)
}

func mutatingWebhookMatches(existing, desired *admissionv1.MutatingWebhookConfiguration, caBundle []byte) bool {
	return mutatingWebhookCABundleValid(existing, caBundle) && equality.Semantic.DeepEqual(existing.Webhooks, desired.Webhooks)
}

func validatingWebhookCABundleValid(vwc *admissionv1.ValidatingWebhookConfiguration, caBundle []byte) bool {
	if len(vwc.Webhooks) == 0 {
		return false
	}
	for i := range vwc.Webhooks {
		if !bytes.Equal(vwc.Webhooks[i].ClientConfig.CABundle, caBundle) {
			return false
		}
	}
	return true
}

func mutatingWebhookCABundleValid(mwc *admissionv1.MutatingWebhookConfiguration, caBundle []byte) bool {
	if len(mwc.Webhooks) == 0 {
		return false
	}
	for i := range mwc.Webhooks {
		if !bytes.Equal(mwc.Webhooks[i].ClientConfig.CABundle, caBundle) {
			return false
		}
	}
	return true
}
