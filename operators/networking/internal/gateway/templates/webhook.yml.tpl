{{- /* # Certificate Issuer */}}
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: {{.NamePrefix}}-selfsigned-issuer
  namespace: {{.Namespace}}
  ownerReferences: {{.OwnerReferences | toYAML |nindent 4 }}
spec:
  selfSigned: {}

---

{{- /* # Certificate */}}
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: {{.NamePrefix}}-webhook-cert
  namespace: {{.Namespace}}
  ownerReferences: {{.OwnerReferences | toYAML |nindent 4 }}
spec:
  secretName: {{.NamePrefix}}-webhook-cert
  dnsNames:
  - {{.ServiceName}}.{{.Namespace}}
  - {{.ServiceName}}.{{.Namespace}}.svc
  issuerRef:
    name: {{.NamePrefix}}-selfsigned-issuer
---

{{- /*#: Service */}}
{{- /* apiVersion: v1 */}}
{{- /* kind: Service */}}
{{- /* metadata: */}}
{{- /*   name: {{.NamePrefix}}-svc */}}
{{- /*   namespace: {{.Namespace}} */}}
{{- /*   ownerReferences: {{.OwnerReferences | toYAML |nindent 4 }} */}}
{{- /* spec: */}}
{{- /*   type: ClusterIP */}}
{{- /*   ports: */}}
{{- /*   - port: 443 */}}
{{- /*     targetPort: 8443 */}}
{{- /*   selector: */}}
{{- /*     name: {{.NamePrefix}}-server */}}

{{- /* --- */}}

{{- /*#: Deployment */}}
{{- /* apiVersion: apps/v1 */}}
{{- /* kind: Deployment */}}
{{- /* metadata: */}}
{{- /*   name: &name {{.NamePrefix}}-server */}}
{{- /*   namespace: {{.Namespace}} */}}
{{- /*   ownerReferences: {{.OwnerReferences | toYAML |nindent 4 }} */}}
{{- /*   labels: &labels */}}
{{- /*     name: *name */}}
{{- /* spec: */}}
{{- /*   replicas: 1 */}}
{{- /*   selector: */}}
{{- /*     matchLabels: *labels */}}
{{- /*   template: */}}
{{- /*     metadata: */}}
{{- /*       labels: *labels */}}
{{- /*     spec: */}}
{{- /*       containers: */}}
{{- /*       - name: socat */}}
{{- /*         image: ghcr.io/kloudlite/hub/socat:latest */}}
{{- /*         command: */}}
{{- /*           - sh */}}
{{- /*           - -c */}}
{{- /*           - |+ */}}
{{- /*             (socat -dd tcp4-listen:8443,fork,reuseaddr tcp4:baby.default.svc.cluster.local:443 2>&1 | grep -iE --line-buffered 'listening|exiting') & */}}
{{- /*             pid="$pid $!" */}}
{{- /**/}}
{{- /*             trap "eval kill -9 $pid || exit 0" EXIT SIGINT SIGTERM */}}
{{- /*             eval wait $pid */}}
{{- /**/}}
{{- /*       - name: webhook-server */}}
{{- /*         image: {{.WebhookServerImage}} */}}
{{- /*         imagePullPolicy: Always */}}
{{- /*         ports: */}}
{{- /*         - containerPort: 8443 */}}
{{- /*         volumeMounts: */}}
{{- /*         - name: tls */}}
{{- /*           mountPath: /tls */}}
{{- /**/}}
{{- /*       volumes: */}}
{{- /*       - name: tls */}}
{{- /*         secret: */}}
{{- /*           secretName: {{.NamePrefix}}-cert */}}

---

{{- /* Webhook */}}
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: {{.NamePrefix}}-webhook
  namespace: {{.Namespace}}
  annotations:
    cert-manager.io/inject-ca-from: {{.Namespace}}/{{.NamePrefix}}-webhook-cert
  ownerReferences: {{.OwnerReferences | toYAML |nindent 4 }}
webhooks:
- name: {{.NamePrefix}}-pod.{{.Namespace}}.webhook.com
  clientConfig:
    service:
      name: {{.ServiceName}}
      namespace: {{.Namespace}}
      path: /mutate/pod
    # caBundle: <CA_BUNDLE> # Replace with the base64 encoded CA certificate
  rules:
  - operations: ["CREATE","DELETE"]
    apiGroups: [""]
    apiVersions: ["v1"]
    resources: ["pods"]
    scope: "Namespaced"

  namespaceSelector:
    matchExpressions:
      - key: kloudlite.io/webhooks.enabled
        operator: In
        values: ["true"]
  admissionReviewVersions: ["v1"]
  sideEffects: None

- name: {{.NamePrefix}}-svc.{{.Namespace}}.webhook.com
  clientConfig:
    service:
      name: {{.ServiceName}}
      namespace: {{.Namespace}}
      path: /mutate/service
    # caBundle: <CA_BUNDLE> # Replace with the base64 encoded CA certificate
  rules:
  - operations: ["CREATE", "DELETE"]
    apiGroups: [""]
    apiVersions: ["v1"]
    resources: ["services"]
    scope: "Namespaced"

  namespaceSelector:
    matchExpressions:
      - key: kloudlite.io/webhooks.enabled
        operator: In
        values: ["true"]
  admissionReviewVersions: ["v1"]
  sideEffects: None
