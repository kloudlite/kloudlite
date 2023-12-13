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
          refName: "mres-{{.Values.managedResources.infraDb}}-creds"
          refKey: URI

        - key: HTTP_PORT
          value: "3000"

        - key: GRPC_PORT
          value: {{.Values.apps.infraApi.configuration.grpcPort | squote}}

        - key: COOKIE_DOMAIN
          value: "{{.Values.cookieDomain}}"

        - key: AUTH_REDIS_HOSTS
          type: secret
          {{- /* refName: "mres-{{.Values.managedResources.authRedis}}" */}}
          refName: "msvc-{{.Values.managedServices.redisSvc}}"
          refKey: HOSTS

        - key: AUTH_REDIS_PASSWORD
          type: secret
          {{- /* refName: "mres-{{.Values.managedResources.authRedis}}" */}}
          {{- /* refKey: PASSWORD */}}
          refName: "msvc-{{.Values.managedServices.redisSvc}}"
          refKey: ROOT_PASSWORD

        - key: AUTH_REDIS_PREFIX
          value: "auth"
          {{- /* type: secret */}}
          {{- /* refName: "mres-{{.Values.managedResources.authRedis}}" */}}
          {{- /* refKey: PREFIX */}}

        - key: AUTH_REDIS_USERNAME
          value: ""
          {{- /* type: secret */}}
          {{- /* refName: "mres-{{.Values.managedResources.authRedis}}" */}}
          {{- /* refKey: USERNAME */}}

        {{- /* - key: KAFKA_BROKERS */}}
        {{- /*   type: secret */}}
        {{- /*   refName: {{.Values.secretNames.redpandaAdminAuthSecret}} */}}
        {{- /*   refKey: KAFKA_BROKERS */}}
        {{- /**/}}
        {{- /* - key: KAFKA_USERNAME */}}
        {{- /*   type: secret */}}
        {{- /*   refName: {{.Values.secretNames.redpandaAdminAuthSecret}} */}}
        {{- /*   refKey: USERNAME */}}
        {{- /**/}}
        {{- /* - key: KAFKA_PASSWORD */}}
        {{- /*   type: secret */}}
        {{- /*   refName: {{.Values.secretNames.redpandaAdminAuthSecret}} */}}
        {{- /*   refKey: PASSWORD */}}
        {{- /**/}}
        {{- /* - key: KAFKA_TOPIC_INFRA_UPDATES */}}
        {{- /*   value: {{ required "env var KAKFA_TOPIC_INFRA_UPDATES must be set" .Values.kafka.topicInfraStatusUpdates }} */}}
        {{- /**/}}
        {{- /* - key: KAFKA_TOPIC_BYOC_CLIENT_UPDATES */}}
        {{- /*   value: {{.Values.kafka.topicBYOCClientUpdates}} */}}
        {{- /**/}}
        {{- /* - key: KAFKA_CONSUMER_GROUP_ID */}}
        {{- /*   value: {{.Values.kafka.consumerGroupId}} */}}

        - key: NATS_URL
          value: "nats://nats.kloudlite.svc.cluster.local:4222"

        - key: NATS_STREAM
          value: resource-sync

        - key: ACCOUNT_COOKIE_NAME
          value: kloudlite-account

        - key: PROVIDER_SECRET_NAMESPACE
          value: {{.Release.Namespace}}

        - key: IAM_GRPC_ADDR
          value: {{.Values.apps.iamApi.name}}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}:{{.Values.apps.iamApi.configuration.grpcPort}}

        - key: MESSAGE_OFFICE_INTERNAL_GRPC_ADDR
          value: {{.Values.apps.messageOfficeApi.name}}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}:{{.Values.apps.messageOfficeApi.configuration.internalGrpcPort}}

        - key: "KAFKA_TOPIC_SEND_MESSAGES_TO_TARGET_WAIT_QUEUE"
          value: "{{.Values.kafka.topicSendMessagesToTargetWaitQueue}}"

        - key: VPN_DEVICES_MAX_OFFSET
          value: {{.Values.apps.consoleApi.configuration.vpnDevicesMaxOffset | squote}}

        - key: VPN_DEVICES_OFFSET_START
          value: {{.Values.apps.consoleApi.configuration.vpnDevicesOffsetStart | squote}}
      
        - key: AWS_ACCESS_KEY
          value: {{.Values.apps.infraApi.configuration.aws.accessKey}}

        - key: AWS_SECRET_KEY
          value: {{.Values.apps.infraApi.configuration.aws.secretKey}}

        - key: AWS_CF_STACK_S3_URL
          value: {{.Values.apps.infraApi.configuration.aws.cloudformation.stackS3URL}}

        - key: AWS_CF_PARAM_TRUSTED_ARN
          value: {{.Values.apps.infraApi.configuration.aws.cloudformation.params.trustedARN}}
        
        - key: AWS_CF_STACK_NAME_PREFIX
          value: {{.Values.apps.infraApi.configuration.aws.cloudformation.stackNamePrefix}}

        - key: AWS_CF_ROLE_NAME_PREFIX
          value: {{.Values.apps.infraApi.configuration.aws.cloudformation.roleNamePrefix}}

        - key: AWS_CF_INSTANCE_PROFILE_NAME_PREFIX
          value: {{.Values.apps.infraApi.configuration.aws.cloudformation.instanceProfileNamePrefix}}

        - key: PUBLIC_DNS_HOST_SUFFIX
          value: {{.Values.baseDomain}}


