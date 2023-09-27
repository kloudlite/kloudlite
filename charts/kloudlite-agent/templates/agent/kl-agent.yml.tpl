{{- if .Values.agent.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.agent.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    checksum/cluster-identity-secret: {{ include (print $.Template.BasePath "/secrets/cluster-identity-secret.yml.tpl") . | sha256sum }}
spec:
  replicas: 1
  serviceAccount: {{include "serviceAccountName" . | quote}}
  services:
    - port: 6000
      targetPort: 6000
      name: grpc
  containers:
    - name: main
      image: {{.Values.agent.image}}
      imagePullPolicy: {{.Values.agent.imagePullPolicy | default .Values.imagePullPolicy }}
      env:
        - key: GRPC_ADDR
          value: {{.Values.messageOfficeGRPCAddr}}

        - key: CLUSTER_TOKEN
          type: secret
          refName: {{.Values.clusterIdentitySecretName}}
          refKey: CLUSTER_TOKEN

        - key: ACCESS_TOKEN
          type: secret
          refName: {{.Values.clusterIdentitySecretName}}
          refKey: ACCESS_TOKEN
          optional: true

        - key: ACCESS_TOKEN_SECRET_NAME
          value: {{.Values.clusterIdentitySecretName}}

        - key: ACCESS_TOKEN_SECRET_NAMESPACE
          value: {{.Release.Namespace}}

        - key: CLUSTER_NAME
          value: {{.Values.clusterName}}

        - key: ACCOUNT_NAME
          value: {{.Values.accountName}}
      
        {{- /* - key: IMAGE_PULL_SECRET_NAME */}}
        {{- /*   value: {{.Values.defaultImagePullSecretName}} */}}
        {{- /**/}}
        {{- /* - key: IMAGE_PULL_SECRET_NAMESPACE */}}
        {{- /*   value: {{.Release.Namespace}} */}}

        - key: VECTOR_PROXY_GRPC_SERVER_ADDR
          value: 0.0.0.0:6000

        - key: RESOURCE_WATCHER_NAME
          value: {{.Values.operators.resourceWatcher.name}}

        - key: RESOURCE_WATCHER_NAMESPACE
          value: {{.Release.Namespace}}

      resourceCpu:
        min: 30m
        max: 50m
      resourceMemory:
        min: 50Mi
        max: 80Mi
{{- end }}
