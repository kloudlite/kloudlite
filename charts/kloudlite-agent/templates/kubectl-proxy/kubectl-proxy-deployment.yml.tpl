{{- $name := "kubectl-proxy" }}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
  annotations:
    checksum/cluster-identity-secret: {{ include (print $.Template.BasePath "/secrets/cluster-identity-secret.yml.tpl") . | sha256sum }}
spec:
  replicas: 1
  selector:
    matchLabels: &labels
      app: {{$name}}
  template:
    metadata:
      labels: *labels
    spec:
      nodeSelector: {{.Values.agent.nodeSelector | default .Values.nodeSelector | toYaml | nindent 8}}
      tolerations: {{.Values.agent.tolerations | default .Values.tolerations | toYaml | nindent 8}}

      serviceAccountName: {{printf "%s-kubectl-proxy" .Release.Name}}
      priorityClassName: kloudlite-critical

      containers:
      - name: main
        image: ghcr.io/kloudlite/hub/kubectl:latest
        imagePullPolicy: IfNotPresent
        command:
          - kubectl
          - proxy
          - -p
          - "8080"
          - --address
          - "0.0.0.0"
          - --accept-hosts
          {{- /* - "^(kubectl-proxy.kloudlite.svc.example-test.local)$" */}}
          - ".*"
        resources:
          limits:
            cpu: 200m
            memory: 200Mi
          requests:
            cpu: 50m
            memory: 50Mi

---

apiVersion: v1
kind: Service
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
spec:
  ports:
  - name: k-proxy
    port: 8080
    protocol: TCP
    targetPort: 8080
  selector:
    app: "{{$name}}"
