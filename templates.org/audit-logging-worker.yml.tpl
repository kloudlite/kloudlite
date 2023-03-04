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
kind: ManagedResource
metadata:
  name: {{.EventsDbName}}
  namespace: {{$namespace}}
  labels:
    kloudlite.io/account-ref:  {{$accountRef}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  inputs:
    resourceName: {{.EventsDbName}}
  msvcRef:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.MongoSvcName}}
  mresKind:
    kind: Database

---
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.AppAuditLoggingWorker}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4 }}
spec:
  region: {{$region}}
  services: []
  containers:
    - name: main
      image: {{.ImageAuditLoggingWorker}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "50m"
        max: "100m"
      resourceMemory:
        min: "100Mi"
        max: "100Mi"
      env:
        - key: KAFKA_BROKERS
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: "KAFKA_BROKERS"

        - key: KAFKA_USERNAME
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: "USERNAME"

        - key: KAFKA_PASSWORD
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: "PASSWORD"

        - key: KAFKA_SUBSCRIPTION_TOPICS
          value: {{.KafkaTopicEvents}}

        - key: KAFKA_CONSUMER_GROUP_ID
          value: "control-plane"

        - key: EVENTS_DB_URI
          type: secret
          refName: {{printf "mres-%s" .EventsDbName}}
          refKey: URI

        - key: EVENTS_DB_NAME
          value: {{.EventsDbName}}
{{end}}
