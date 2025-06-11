apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Values.agent.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    checksum/cluster-identity-secret: {{ include (print $.Template.BasePath "/secrets/cluster-identity-secret.yml.tpl") . | sha256sum }}
    vector.dev/exclude: "true" # to exclude pods from being monitored by vector
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{.Values.agent.name}}
  template:
    metadata:
      labels:
        app: {{.Values.agent.name}}
        vector.dev/exclude: "true" # to exclude pods from being monitored by vector
    spec:
      nodeSelector: {{.Values.agent.nodeSelector | default .Values.nodeSelector | toYaml | nindent 8}}
      tolerations: {{.Values.agent.tolerations | default .Values.tolerations | toYaml | nindent 8}}

      serviceAccountName: {{include "serviceAccountName" . | quote}}
      priorityClassName: kloudlite-critical

      containers:
      - name: main
        {{- $imageTag := .Values.agent.image.tag | default (include "image-tag" .) }}
        image: {{.Values.agent.image.repository}}:{{$imageTag}}
        imagePullPolicy: {{ .Values.agent.image.pullPolicy | default (.Values.imagePullPolicy | default (include "image-pull-policy" $imageTag)) }}
        env:
          - name: GRPC_ADDR
            value: {{.Values.messageOfficeGRPCAddr}}

          - name: CLUSTER_TOKEN
            valueFrom:
              secretKeyRef:
                key: CLUSTER_TOKEN
                name: {{.Values.clusterIdentitySecretName}}
                optional: false

          - name: ACCESS_TOKEN
            valueFrom:
              secretKeyRef:
                key: ACCESS_TOKEN
                name: {{.Values.clusterIdentitySecretName}}
                optional: true

          - name: ACCESS_TOKEN_SECRET_NAME
            value: {{.Values.clusterIdentitySecretName}}

          - name: ACCESS_TOKEN_SECRET_NAMESPACE
            value: {{.Release.Namespace}}

          - name: VECTOR_PROXY_GRPC_SERVER_ADDR
            value: 0.0.0.0:6000

          - name: RESOURCE_WATCHER_NAME
            value: {{.Values.agentOperator.name}}

          - name: RESOURCE_WATCHER_NAMESPACE
            value: {{.Release.Namespace}}

        resources: {{.Values.agent.resources | toYaml | nindent 10}}

---

apiVersion: v1
kind: Service
metadata:
  name: {{.Values.agent.name}}
  namespace: {{.Release.Namespace}}
spec:
  ports:
  - name: "vector-grpc-proxy"
    port: 6000
    protocol: TCP
    targetPort: 6000
  selector:
    app: "{{.Values.agent.name}}"
---
