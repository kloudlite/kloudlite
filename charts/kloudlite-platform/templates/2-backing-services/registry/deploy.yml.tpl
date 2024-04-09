{{- $name := .Release.Name -}}
{{- $configName := .Release.Name -}}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: distribution
  labels:
    app: {{ $name }}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{ $name }}
  template:
    metadata:
      labels:
        app: {{ $name }}
    spec:
      tolerations: {{.Values.nodepools.stateless.tolerations | toYaml | nindent 8 }}
      nodeSelector: {{.Values.nodepools.stateless.labels | toYaml | nindent 8 }}
      containers:
      - name: registry-container
        image: registry:2
{{/*        resources: {{ .Values.registry.resources | toYaml | nindent 10 }}*/}}
        ports:
        - containerPort: 5000
          name: http
          protocol: TCP

        volumeMounts:
        - name: config-volume
          mountPath: /etc/docker/registry
        - name: config-volume
          mountPath: /etc/docker/registry

      volumes:
      - name: config-volume
        configMap:
          name: {{ $configName }}
