{{- $namespace := get . "namespace" -}}
{{- $svcAccount := get . "svc-account" -}}
{{- $sharedConstants := get . "shared-constants" -}}

{{- $ownerRefs := get . "owner-refs" | default list -}}
{{- $accountRef := get . "account-ref" | default "kl-core" -}}
{{- $region := get . "region" | default "master" -}}
{{- $imagePullPolicy := get . "image-pull-policy" | default "Always" -}}

{{- $nodeSelector := get . "node-selector" | default dict -}}
{{- $tolerations := get . "tolerations" | default list -}}

{{ with $sharedConstants}}
{{/*gotype: github.com/kloudlite/operator/apis/cluster-setup/v1.SharedConstants*/}}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.AppWebhooksApi}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  region: {{$region}}
  nodeSelector: {{$nodeSelector | toYAML | nindent 4}}
  tolerations: {{$tolerations | toYAML | nindent 4 }}
  services:
    - port: 80
      targetPort: 3000
      type: tcp
  containers:
    - name: main
      image: {{.ImageWebhooksApi}}
      imagePullPolicy: {{$imagePullPolicy}}
      env:
        - key: CI_ADDR
          value: "{{.AppCiApi}}.{{$namespace}}.svc.cluster.local:80"

        - key: GITHUB_AUTHZ_SECRET
          type: secret
          refName: {{.WebhookAuthzSecretName}}
          refKey: GITHUB_AUTHZ_SECRET

        - key: GITLAB_AUTHZ_SECRET
          type: secret
          refName: {{.WebhookAuthzSecretName}}
          refKey: GITLAB_AUTHZ_SECRET

        - key: HARBOR_AUTHZ_SECRET
          type: secret
          refName: {{.WebhookAuthzSecretName}}
          refKey: HARBOR_AUTHZ_SECRET

        - key: KL_HOOK_TRIGGER_AUTHZ_SECRET
          type: secret
          refName: {{.WebhookAuthzSecretName}}
          refKey: KL_HOOK_TRIGGER_AUTHZ_SECRET

        - key: HTTP_PORT
          value: "3000"

        - key: KAFKA_BROKERS
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: KAFKA_BROKERS

        - key: GIT_WEBHOOKS_TOPIC
          value: {{.KafkaTopicGitWebhooks}}

        - key: HARBOR_WEBHOOK_TOPIC
          value: {{.KafkaTopicHarborWebhooks}}

        - key: KAFKA_USERNAME
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: USERNAME

        - key: KAFKA_PASSWORD
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: PASSWORD
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"

---

apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.AppWebhooksApi}}
  namespace: {{$namespace}}
  labels:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  domains:
    - webhooks.{{.SubDomain}}
  https:
    enabled: true
  routes:
    - app: {{.AppWebhooksApi}}
      path: /
      port: 80
{{end}}
