package services

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"go.uber.org/zap"
	admissionv1 "k8s.io/api/admissionregistration/v1"
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
	wi.logger.Info("Installing webhook configurations...")

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

	wi.logger.Info("Webhook configurations installed successfully")
	return nil
}

func (wi *WebhookInstaller) createOrUpdateValidatingWebhook(ctx context.Context, vwc *admissionv1.ValidatingWebhookConfiguration) error {
	existing := &admissionv1.ValidatingWebhookConfiguration{}
	err := wi.k8sClient.Get(ctx, client.ObjectKey{Name: vwc.Name}, existing)

	if errors.IsNotFound(err) {
		wi.logger.Info("Creating ValidatingWebhookConfiguration", zap.String("name", vwc.Name))
		return wi.k8sClient.Create(ctx, vwc)
	} else if err != nil {
		return err
	}

	// Update existing
	wi.logger.Info("Updating ValidatingWebhookConfiguration", zap.String("name", vwc.Name))
	vwc.ResourceVersion = existing.ResourceVersion
	return wi.k8sClient.Update(ctx, vwc)
}

func (wi *WebhookInstaller) createOrUpdateMutatingWebhook(ctx context.Context, mwc *admissionv1.MutatingWebhookConfiguration) error {
	existing := &admissionv1.MutatingWebhookConfiguration{}
	err := wi.k8sClient.Get(ctx, client.ObjectKey{Name: mwc.Name}, existing)

	if errors.IsNotFound(err) {
		wi.logger.Info("Creating MutatingWebhookConfiguration", zap.String("name", mwc.Name))
		return wi.k8sClient.Create(ctx, mwc)
	} else if err != nil {
		return err
	}

	// Update existing
	wi.logger.Info("Updating MutatingWebhookConfiguration", zap.String("name", mwc.Name))
	mwc.ResourceVersion = existing.ResourceVersion
	return wi.k8sClient.Update(ctx, mwc)
}
