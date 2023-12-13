apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.messageOfficeApi.name}}
  namespace: {{.Release.Namespace}}
spec:
  {{/* serviceAccount: {{.Values.normalSvcAccount}} */}}
  serviceAccount: {{.Values.clusterSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 80
      targetPort: {{.Values.apps.messageOfficeApi.configuration.httpPort}}
      name: http
      type: tcp

    - port: {{.Values.apps.messageOfficeApi.configuration.externalGrpcPort}}
      targetPort: {{.Values.apps.messageOfficeApi.configuration.externalGrpcPort}}
      name: grpc
      type: tcp

    - port: {{.Values.apps.messageOfficeApi.configuration.internalGrpcPort}}
      targetPort: {{.Values.apps.messageOfficeApi.configuration.internalGrpcPort}}
      name: internal-grpc
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.messageOfficeApi.image}}
      imagePullPolicy: {{.Values.apps.messageOfficeApi.imagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "100m"
        max: "150m"
      resourceMemory:
        min: "100Mi"
        max: "150Mi"
      env:
        - key: HTTP_PORT
          value: {{.Values.apps.messageOfficeApi.configuration.httpPort | squote}}

        - key: EXTERNAL_GRPC_PORT
          value: {{.Values.apps.messageOfficeApi.configuration.externalGrpcPort | squote}}

        - key: INTERNAL_GRPC_PORT
          value: {{.Values.apps.messageOfficeApi.configuration.internalGrpcPort | squote}}

        - key: DB_URI
          type: secret
          refName: "mres-{{.Values.managedResources.messageOfficeDb}}-creds"
          refKey: URI

        - key: DB_NAME
          value: {{.Values.managedResources.messageOfficeDb}}

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

        {{- /* - key: KAFKA_TOPIC_STATUS_UPDATES */}}
        {{- /*   value: {{.Values.kafka.topicStatusUpdates}} */}}
        {{- /**/}}
        {{- /* - key: KAFKA_TOPIC_INFRA_UPDATES */}}
        {{- /*   value: {{.Values.kafka.topicInfraStatusUpdates}} */}}
        {{- /**/}}
        {{- /* - key: KAFKA_TOPIC_ERROR_ON_APPLY */}}
        {{- /*   value: {{.Values.kafka.topicErrorOnApply}} */}}
        {{- /**/}}
        {{- /* - key: KAFKA_TOPIC_CLUSTER_UPDATES */}}
        {{- /*   value: {{.Values.kafka.topicClusterUpdates}} */}}
        {{- /**/}}
        {{- /* - key: KAFKA_CONSUMER_GROUP */}}
        {{- /*   value: {{.Values.kafka.consumerGroupId}} */}}
        {{- /**/}}
        {{- /* - key: KAFKA_BROKERS */}}
        {{- /*   type: secret */}}
        {{- /*   refName: "{{.Values.secretNames.redpandaAdminAuthSecret}}" */}}
        {{- /*   refKey: KAFKA_BROKERS */}}
        {{- /**/}}
        {{- /* - key: KAFKA_SASL_USERNAME */}}
        {{- /*   type: secret */}}
        {{- /*   refName: "{{.Values.secretNames.redpandaAdminAuthSecret}}" */}}
        {{- /*   refKey: USERNAME */}}
        {{- /**/}}
        {{- /* - key: KAFKA_SASL_PASSWORD */}}
        {{- /*   type: secret */}}
        {{- /*   refName: "{{.Values.secretNames.redpandaAdminAuthSecret}}" */}}
        {{- /*   refKey: PASSWORD */}}

        - key: NATS_URL
          value: "nats://nats.kloudlite.svc.cluster.local:4222"

        - key: NATS_STREAM
          value: "resource-sync"

        - key: VECTOR_GRPC_ADDR
          value: {{printf "%s:6000" (include "vector.name" .) | quote}}

        - key: TOKEN_HASHING_SECRET
          value: {{.Values.apps.messageOfficeApi.configuration.tokenHashingSecret | squote}}

