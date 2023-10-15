apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.infraApi.name}}
  namespace: {{.Release.Namespace}}
spec:
  region: {{.Values.region | default ""}}
  serviceAccount: {{.Values.clusterSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 4 }}

  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.infraApi.image}}
      imagePullPolicy: {{.Values.apps.infraApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      
      resourceCpu:
        min: "50m"
        max: "100m"
      resourceMemory:
        min: "50Mi"
        max: "100Mi"

      env:
        {{- /* - key: FINANCE_GRPC_ADDR */}}
        {{- /*   value: http://{{.Values.apps.financeApi.name}}:3001 */}}
        - key: ACCOUNTS_GRPC_ADDR
          value: {{.Values.apps.accountsApi.name}}:{{.Values.apps.accountsApi.configuration.grpcPort}}

        - key: INFRA_DB_NAME
          value: {{.Values.managedResources.infraDb}}

        - key: INFRA_DB_URI
          type: secret
          refName: "mres-{{.Values.managedResources.infraDb}}"
          refKey: URI

        - key: HTTP_PORT
          value: "3000"

        - key: GRPC_PORT
          value: {{.Values.apps.infraApi.configuration.grpcPort | squote}}

        - key: COOKIE_DOMAIN
          value: "{{.Values.cookieDomain}}"

        - key: AUTH_REDIS_HOSTS
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: HOSTS

        - key: AUTH_REDIS_PASSWORD
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: PASSWORD

        - key: AUTH_REDIS_PREFIX
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: PREFIX

        - key: AUTH_REDIS_USER_NAME
          type: secret
          refName: "mres-{{.Values.managedResources.authRedis}}"
          refKey: USERNAME

        - key: KAFKA_BROKERS
          type: secret
          refName: {{.Values.secretNames.redpandaAdminAuthSecret}}
          refKey: KAFKA_BROKERS

        - key: KAFKA_USERNAME
          type: secret
          refName: {{.Values.secretNames.redpandaAdminAuthSecret}}
          refKey: USERNAME

        - key: KAFKA_PASSWORD
          type: secret
          refName: {{.Values.secretNames.redpandaAdminAuthSecret}}
          refKey: PASSWORD

        - key: KAFKA_TOPIC_INFRA_UPDATES
          value: {{.Values.kafka.topicinfraStatusUpdates}}

        - key: KAFKA_TOPIC_BYOC_CLIENT_UPDATES
          value: {{.Values.kafka.topicBYOCClientUpdates}}

        - key: KAFKA_CONSUMER_GROUP_ID
          value: {{.Values.kafka.consumerGroupId}}

        - key: ACCOUNT_COOKIE_NAME
          value: kloudlite-account

        - key: PROVIDER_SECRET_NAMESPACE
          value: {{.Release.Namespace}}

        - key: IAM_GRPC_ADDR
          value: {{.Values.apps.iamApi.name}}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}:{{.Values.apps.iamApi.configuration.grpcPort}}

        - key: MESSAGE_OFFICE_INTERNAL_GRPC_ADDR
          value: {{.Values.apps.messageOfficeApi.name}}.{{.Release.Namespace}}.{{.Values.clusterInternalDNS}}:{{.Values.apps.messageOfficeApi.configuration.internalGrpcPort}}

