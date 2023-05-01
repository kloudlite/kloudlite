{{- if .Values.agent.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.agent.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    "helm.sh/hook": post-install
    "helm.sh/hook-weight": "99"

spec:
  replicas: 1
  {{/* accountName: {{.Values.accountName}} */}}
  region: {{.Values.region}}
  serviceAccount: {{.Values.svcAccountName}}
  containers:
    - name: main
      image: {{.Values.agent.image}}
      imagePullPolicy: {{.Values.agent.imagePullPolicy | default .Values.imagePullPolicy }}
      env:
        {{/* - key: KAFKA_BROKERS */}}
        {{/*   type: secret */}}
        {{/*   refName: {{.Values.kafka.secretName}} */}}
        {{/*   refKey: KAFKA_BROKERS */}}
        {{/**/}}
        {{/* - key: KAFKA_SASL_USER */}}
        {{/*   type: secret */}}
        {{/*   refName: {{.Values.kafka.secretName}} */}}
        {{/*   refKey: KAFKA_SASL_USERNAME */}}
        {{/**/}}
        {{/* - key: KAFKA_SASL_PASSWORD */}}
        {{/*   type: secret */}}
        {{/*   refName: {{.Values.kafka.secretName}} */}}
        {{/*   refKey: KAFKA_SASL_PASSWORD */}}
        {{/**/}}
        {{/* - key: KAFKA_INCOMING_TOPIC */}}
        {{/*   value: {{.Values.kafka.topicClusterIncoming}} */}}
        {{/**/}}
        {{/* - key: KAFKA_CONSUMER_GROUP_ID */}}
        {{/*   value: "{{.Values.kafka.consumerGroupClusterIncoming}}" */}}
        {{/**/}}
        {{/* - key: KAFKA_ERROR_ON_APPLY_TOPIC */}}
        {{/*   value: {{.Values.kafka.topicErrorOnApply}} */}}

        - key: GRPC_ADDR
          value: {{.Values.kloudliteLinks.messageOfficeGRPCAddr}}

        - key: CLUSTER_TOKEN
          type: secret
          refName: {{.Values.clusterIdentitySecretName}}
          refKey: CLUSTER_TOKEN
          optional: true

        - key: ACCESS_TOKEN
          type: secret
          refName: {{.Values.clusterIdentitySecretName}}
          refKey: ACCESS_TOKEN
          optional: true

        - key: ACCESS_TOKEN_SECRET_NAME
          value: {{.Values.clusterIdentitySecretName}}

        - key: ACCESS_TOKEN_SECRET_NAMESPACE
          value: {{.Release.Namespace}}

      resourceCpu:
        min: 30m
        max: 50m
      resourceMemory:
        min: 30Mi
        max: 50Mi
{{- end }}
