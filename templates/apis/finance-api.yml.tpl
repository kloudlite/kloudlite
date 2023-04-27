apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.financeApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region | default ""}}
  serviceAccount: {{.Values.clusterSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}
  
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
      image: {{.Values.apps.financeApi.image}}
      imagePullPolicy: {{.Values.apps.financeApi.ImagePullPolicy | default .Values.imagePullPolicy }}

      resourceCpu:
        min: "50m"
        max: "100m"
      resourceMemory:
        min: "80Mi"
        max: "100Mi"
      env:
        - key: HTTP_PORT
          value: "3000"

        - key: GRPC_PORT
          value: "3001"

        - key: CONSOLE_SERVICE
          value: "{{.Values.apps.consoleApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001"

        - key: AUTH_SERVICE
          value: "{{.Values.apps.authApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001"
        - key: IAM_SERVICE
          value: "{{.Values.apps.iamApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001"

        {{/* - key: CI_SERVICE */}}
        {{/*   value: "{{.Values.apps.ciApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001" */}}

        - key: CONTAINER_REGISTRY_SERVICE
          value: "{{.Values.apps.containerRegistryApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001"

        - key: COOKIE_DOMAIN
          value: "{{.Values.cookieDomain}}"

        - key: MONGO_DB_NAME
          value: {{.Values.managedResources.financeDb}}

        - key: COMMS_SERVICE
          value: {{.Values.apps.commsApi.name}}:3001

        - key: REDIS_HOSTS
          type: secret
          refName: mres-{{.Values.managedResources.financeRedis}}
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: mres-{{.Values.managedResources.financeRedis}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: mres-{{.Values.managedResources.financeRedis}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: mres-{{.Values.managedResources.financeRedis}}
          refKey: USERNAME

        - key: REDIS_AUTH_HOSTS
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: HOSTS

        - key: REDIS_AUTH_PASSWORD
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: PASSWORD

        - key: REDIS_AUTH_PREFIX
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: PREFIX

        - key: REDIS_AUTH_USERNAME
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: USERNAME

        - key: MONGO_URI
          type: secret
          refName: mres-{{.Values.managedResources.financeDb}}
          refKey: URI

        - key: INVENTORY_PATH
          value: /finance/inventory

        - key: WORKLOAD_KAFKA_BROKERS
          type: secret
          refName: {{.Values.secrets.names.redpandaAdminAuthSecret}}
          refKey: KAFKA_BROKERS

        - key: KAFKA_WORKLOAD_FINANCE_TOPIC
          value: {{.Values.kafka.topicBilling}}

        - key: KAFKA_USERNAME
          type: secret
          refName: {{.Values.secrets.names.redpandaAdminAuthSecret}}
          refKey: USERNAME

        - key: KAFKA_PASSWORD
          type: secret
          refName: {{.Values.secrets.names.redpandaAdminAuthSecret}}
          refKey: PASSWORD

        - key: CLUSTER_CONFIGS_PATH
          value: /tmp/k8s

        - key: CURR_CLUSTER_CONFIG_NAMESPACE
          value: {{.Release.Namespace}}

        - key: CURR_CLUSTER_CONFIG_NAME
          value: "current-cluster"

        - key: CURR_CLUSTER_CONFIG_CLUSTER_ID_KEY
          value: "clusterId"

        - key: STRIPE_PUBLIC_KEY
          type: secret
          refName: {{.Values.secrets.names.stripeSecret}}
          refKey: PUBLIC_KEY

        - key: STRIPE_SECRET_KEY
          type: secret
          refName: {{.Values.secrets.names.stripeSecret}}
          refKey: SECRET_KEY

      volumes:
        - mountPath: /finance/inventory
          type: config
          refName: "{{.Values.apps.financeApi.name}}-inventory-config"

---

# inventory configmap
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.Values.apps.financeApi.name}}-inventory-config
  namespace: {{.Release.Namespace}}
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
