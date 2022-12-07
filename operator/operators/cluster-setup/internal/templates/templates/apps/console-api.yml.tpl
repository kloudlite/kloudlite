{{- $namespace := get . "namespace" -}}
{{- $accountRef := get . "account-ref" | default "kl-core" -}}
{{- $region := get . "region" | default "master" -}}
{{- $image := get . "image" -}}
{{- $imagePullPolicy := get . "image-pull-policy" | default "Always" -}}
{{- $baseDomain := get . "base-domain" -}}

{{/*{{- $kafkaBrokers := get . "kafka-brokers" -}}*/}}

{{/*constants*/}}
{{- $consoleRedis := "console-redis" -}}
{{- $authRedis := "auth-redis" -}}
{{- $consoleDb := "console-db" -}}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: console-api
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
spec:
  region: blr1
  serviceAccount: kloudlite-cluster-svc-account
  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp

    - port: 3001
      targetPort: 3001
      name: grpc
      type: tcp

    - port: 3002
      targetPort: 8192
      name: log
      type: tcp

    - port: 9191
      targetPort: 9191
      name: metrics
      type: tcp

  containers:
    - name: main
      image: {{$image}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"

      env:
        - key: MONGO_DB_NAME
          value: console-db

        - key: REDIS_HOSTS
          type: secret
          refName: "{{$consoleRedis}}"
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: {{$consoleRedis}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: {{$consoleRedis}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: {{$consoleRedis}}
          refKey: USERNAME

        - key: REDIS_AUTH_HOSTS
          type: secret
          refName: {{$authRedis}}
          refKey: HOSTS

        - key: REDIS_AUTH_PASSWORD
          type: secret
          refName: {{$authRedis}}
          refKey: PASSWORD

        - key: REDIS_AUTH_PREFIX
          type: secret
          refName: {{$authRedis}}
          refKey: PREFIX

        - key: REDIS_AUTH_USERNAME
          type: secret
          refName: {{$authRedis}}
          refKey: USERNAME

        - key: MONGO_URI
          type: secret
          refName: {{$consoleDb}}
          refKey: URI

        - key: MANAGED_TEMPLATES_PATH
          value: /console.d/templates/managed-svc-templates.yml

        - key: IAM_SERVICE
          value: iam-api.{{$namespace}}.svc.cluster.local:3001

        - key: FINANCE_SERVICE
          value: finance-api.{{$namespace}}.svc.cluster.local:3001

        - key: AUTH_SERVICE
          value: auth-api.{{$namespace}}.svc.cluster.local:3001

        - key: CI_SERVICE
          value: ci-api.{{$namespace}}.svc.cluster.local:3001

        - key: IAM_SERVICE
          value: iam-api.{{$namespace}}.svc.cluster.local:3001

        - key: IAM_SERVICE
          value: iam-api.{{$namespace}}.svc.cluster.local:3001

        - key: DNS_SERVICE
          value: dns-api.{{$namespace}}.svc.cluster.local:3001

        - key: KAFKA_BOOTSTRAP_SERVERS
          value: ${KAFKA_BROKERS}

        - key: KAFKA_WORKLOAD_STATUS_TOPIC
          value: ${REDPANDA_TOPIC_STATUS_UPDATES}

        - key: KAFKA_GROUP_ID
          value: control-plane

        - key: COMPUTE_PLANS_PATH
          value: /console.d/templates/compute-plans.yaml

        - key: INVENTORY_PATH
          value: /console.d/templates

        - key: JSEVAL_SERVICE
          value: js-eval-api:3001

        - key: KUBE_API_ADDRESS
          value: http://127.0.0.1:2999

        - key: KAFKA_USERNAME
          value: ${KAFKA_SASL_USERNAME}

        - key: KAFKA_PASSWORD
          value: ${KAFKA_SASL_PASSWORD}

        - key: CLUSTER_CONFIGS_PATH
          value: /tmp/k8s

      envFrom:
        - type: secret
          refName: console-env

      volumes:
        - mountPath: /console.d/templates
          type: config
          refName: console-managed-svc-template

        - mountPath: /tmp/k8s
          type: secret
          refName: aggregated-kubeconfigs
---
apiVersion: v1
kind: Secret
metadata:
  name: console-env
  namespace: {{$namespace}}
stringData:
  IMAGE_REGISTRY_PREFIX: registry.kloudlite.io
  LOG_PORT: "8192"

  PORT: "3000"
  GRPC_PORT: "3001"

  NOTIFIER_URL: "http://socket-web.{{$namespace}}.svc.cluster.local:3001"

  # LOKI_URL: loki-external.kl-01.$DOMAIN_1
  LOKI_URL: "loki-external.REPLACE_ME.clusters.{{$baseDomain}}"
  LOKI_AUTH_PASSWORD: "$LOKI_EXTERNAL_PASSWORD"

  METRICS_HTTP_PORT: "9191"
  METRICS_HTTP_CORS: "https://console.{{$baseDomain}}"
  PROMETHEUS_ENDPOINT: "https://prom-external.REPLACE_ME.clusters.{{$baseDomain}}"
  PROMETHEUS_BASIC_AUTH_PASSWORD: "$PROM_EXTERNAL_PASSWORD"

  LOG_SERVER_PORT: "8192"
  COOKIE_DOMAIN: ".{{$baseDomain}}"

  # Kafka
  KAFKA_USERNAME: "admin"
  KAFKA_PASSWORD: ""
  KAFKA_BOOTSTRAP_SERVERS: ""
  KAFKA_WORKLOAD_STATUS_TOPIC: "kl-status-updates"
  KAFKA_GROUP_ID: "control-plane"

---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: console-api
  namespace: {{$namespace}}
  labels:
    kloudlite.io/account-ref: {{$accountRef}}
spec:
  domains:
    - "logs.{{$baseDomain}}"
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: console-api
      path: /metrics
      port: 9191

    - app: console-api
      path: /
      port: 3002
