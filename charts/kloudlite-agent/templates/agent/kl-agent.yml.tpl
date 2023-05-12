{{- if .Values.agent.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.agent.name}}
  namespace: {{.Release.Namespace}}
spec:
  replicas: 1
  serviceAccount: {{.Values.svcAccountName}}
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

        - key: CLUSTER_NAME
          value: {{.Values.clusterName}}

        - key: ACCOUNT_NAME
          value: {{.Values.accountName}}

      resourceCpu:
        min: 30m
        max: 50m
      resourceMemory:
        min: 30Mi
        max: 50Mi
{{- end }}
