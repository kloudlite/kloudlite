---
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: {{.NamePrefix}}-webhook
  namespace: {{.Namespace}}
  annotations:
    {{- /* cert-manager.io/inject-ca-from: {{.Namespace}}/{{.NamePrefix}}-webhook-cert */}}
  ownerReferences: {{.OwnerReferences | toYAML |nindent 4 }}
webhooks:
- name: {{.NamePrefix}}-pod.{{.Namespace}}.webhook.com
  clientConfig:
    service:
      name: {{.ServiceName}}
      namespace: {{.Namespace}}
      path: /mutate/pod
    # caBundle: <CA_BUNDLE> # Replace with the base64 encoded CA certificate
    caBundle: {{.WebhookServerCertCABundle | b64enc}}
  rules:
  - operations: ["CREATE","DELETE"]
    apiGroups: [""]
    apiVersions: ["v1"]
    resources: ["pods"]
    scope: "Namespaced"

  namespaceSelector:
    matchExpressions:
      {{- range $key, $value := .WebhookNamespaceSelector }}
      - key: {{$key}}
        operator: In
        values: [{{$value | squote}}]
      {{- end }}

      {{- /* - key: {{.WebhookNamespaceSelectorKey}} */}}
      {{- /*   operator: In */}}
      {{- /*   values: ["{{.WebhookNamespaceSelectorValue}}"] */}}
  admissionReviewVersions: ["v1"]
  sideEffects: None

- name: {{.NamePrefix}}-svc.{{.Namespace}}.webhook.com
  clientConfig:
    service:
      name: {{.ServiceName}}
      namespace: {{.Namespace}}
      path: /mutate/service
    # caBundle: <CA_BUNDLE> # Replace with the base64 encoded CA certificate
    caBundle: {{.WebhookServerCertCABundle | b64enc}}
  rules:
  - operations: ["CREATE", "UPDATE", "DELETE"]
    apiGroups: [""]
    apiVersions: ["v1"]
    resources: ["services"]
    scope: "Namespaced"

  namespaceSelector:
    matchExpressions:
      {{- range $key, $value := .WebhookNamespaceSelector }}
      - key: {{$key}}
        operator: In
        values: [{{$value | squote}}]
      {{- end }}
      {{- /* - key: {{.WebhookNamespaceSelectorKey}} */}}
      {{- /*   operator: In */}}
      {{- /*   values: ["{{.WebhookNamespaceSelectorValue}}"] */}}
  admissionReviewVersions: ["v1"]
  sideEffects: None
