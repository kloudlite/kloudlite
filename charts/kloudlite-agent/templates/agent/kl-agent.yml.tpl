{{- if .Values.agent.enabled }}

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
        imagePullPolicy: {{ include "image-pull-policy" $imageTag }}
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

          - name: CLUSTER_NAME
            valueFrom:
              secretKeyRef:
                key: CLUSTER_NAME
                name: {{.Values.clusterIdentitySecretName}}

          - name: ACCOUNT_NAME
            valueFrom:
              secretKeyRef:
                key: ACCOUNT_NAME
                name: {{.Values.clusterIdentitySecretName}}

          - name: VECTOR_PROXY_GRPC_SERVER_ADDR
            value: 0.0.0.0:6000

          - name: RESOURCE_WATCHER_NAME
            value: {{.Values.operators.agentOperator.name}}

          - name: RESOURCE_WATCHER_NAMESPACE
            value: {{.Release.Namespace}}

        resources:
          limits:
            cpu: 100m
            memory: 200Mi
          requests:
            cpu: 50m
            memory: 100Mi

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

{{- end  }}
