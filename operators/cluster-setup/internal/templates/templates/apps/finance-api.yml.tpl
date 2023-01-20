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
  name: {{.AppFinanceApi}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  region: {{$region}}
  nodeSelector: {{$nodeSelector | toYAML| nindent 4}}
  tolerations: {{$tolerations | toYAML | nindent 4}}
  serviceAccount: {{$svcAccount}}
  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
    - port: 3001
      targetPort: 3001
      name: grpc
      type: tcp
  containers:
    - name: main
      image: {{.ImageFinanceApi}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"
      env:
        - key: MONGO_DB_NAME
          value: {{.FinanceDbName}}

        - key: COMMS_SERVICE
          value: {{.AppCommsApi}}:3001

        - key: REDIS_HOSTS
          type: secret
          refName: mres-{{.FinanceRedisName}}
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: mres-{{.FinanceRedisName}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: mres-{{.FinanceRedisName}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: mres-{{.FinanceRedisName}}
          refKey: USERNAME

        - key: REDIS_AUTH_HOSTS
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: HOSTS

        - key: REDIS_AUTH_PASSWORD
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: PASSWORD

        - key: REDIS_AUTH_PREFIX
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: PREFIX

        - key: REDIS_AUTH_USERNAME
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: USERNAME

        - key: MONGO_URI
          type: secret
          refName: mres-{{.FinanceDbName}}
          refKey: URI

        - key: INVENTORY_PATH
          value: /finance/inventory

        - key: WORKLOAD_KAFKA_BROKERS
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: KAFKA_BROKERS

        - key: KAFKA_WORKLOAD_FINANCE_TOPIC
          value: $kafkaBillingTopic

        - key: KAFKA_USERNAME
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: USERNAME

        - key: KAFKA_PASSWORD
          type: secret
          refName: {{.RedpandaAdminSecretName}}
          refKey: PASSWORD

        - key: CLUSTER_CONFIGS_PATH
          value: /tmp/k8s

        - key: CURR_CLUSTER_CONFIG_NAMESPACE
          value: {{$namespace}}

        - key: CURR_CLUSTER_CONFIG_NAME
          value: "current-cluster"

        - key: CURR_CLUSTER_CONFIG_CLUSTER_ID_KEY
          value: "clusterId"

        - key: STRIPE_PUBLIC_KEY
          type: secret
          refName: {{.StripeSecretName}}
          refKey: STRIPE_PUBLIC_KEY

        - key: STRIPE_SECRET_KEY
          type: secret
          refName: {{.StripeSecretName}}
          refKey: STRIPE_SECRET_KEY

      envFrom:
        - type: secret
          refName: finance-env

      volumes:
        - mountPath: /finance/inventory
          type: config
          refName: finance-inventory-config
{{/*        - mountPath: /tmp/k8s*/}}
{{/*          type: secret*/}}
{{/*          refName: kl-01-kubeconfig*/}}
{{/*          items:*/}}
{{/*            - key: kubeconfig*/}}
{{/*              fileName: kl-01*/}}

---
# finance-env secret
apiVersion: v1
kind: Secret
metadata:
  name: finance-env
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
stringData:
  PORT: "3000"
  GRPC_PORT: "3001"
  CONSOLE_SERVICE: "{{.AppConsoleApi}}.{{$namespace}}.svc.cluster.local:3001"
  AUTH_SERVICE: "{{.AppAuthApi}}.{{$namespace}}.svc.cluster.local:3001"
  IAM_SERVICE: "{{.AppIAMApi}}.{{$namespace}}.svc.cluster.local:3001"
  CI_SERVICE: "{{.AppCiApi}}.{{$namespace}}.svc.cluster.local:3001"

  COOKIE_DOMAIN: ".{{.CookieDomain}}"

{{/*  STRIPE_PUBLIC_KEY: ***REMOVED****/}}
{{/*  STRIPE_SECRET_KEY: ***REMOVED****/}}


  # STRIPE_PUBLIC_KEY: ***REMOVED***
  # STRIPE_SECRET_KEY: ***REMOVED***

---
# inventory configmap
apiVersion: v1
kind: ConfigMap
metadata:
  name: finance-inventory-config
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
data:
  block-storage.yaml: |+
    - name: BlockStorage
      pricePerGB: 0.1

  ci.yaml: |+

  compute.yaml: |+
    - name: Basic
      sharedPrice: 7.5
      dedicatedPrice: 14

    - name: General
      sharedPrice: 11.5
      dedicatedPrice: 21

    - name: HighMemory
      sharedPrice: 15.5
      dedicatedPrice: 28

  lambda.yaml: |+
    - name: Default
      pricePerGBHr: 0.05
      freeTier: 1000

{{end}}
