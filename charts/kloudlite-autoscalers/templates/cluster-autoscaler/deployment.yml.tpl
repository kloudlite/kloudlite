{{- if .Values.clusterAutoscaler.enabled }}

{{- $name := "kl-cluster-autoscaler" }} 

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
  labels: &labels
    app: {{$name}}
spec:
  selector:
    matchLabels: *labels
  replicas: 1
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: manager
      labels: *labels
    spec:
      securityContext:
        runAsNonRoot: true

      tolerations: {{.Values.clusterAutoscaler.tolerations | default list | toYaml | nindent 8}}
      nodeSelector: {{.Values.clusterAutoscaler.nodeSelector | default dict | toYaml | nindent 8}}

      containers:
        - command:
            - /cluster-autoscaler
          args:
            - --cloud-provider=kloudlite
            - --logtostderr=true
            - --stderrthreshold=info
            - scale-down-unneeded-time=10m
          image: {{.Values.clusterAutoscaler.image.repository}}:{{.Values.clusterAutoscaler.image.tag | default .Values.defaults.imageTag  | default .Chart.AppVersion }}
          imagePullPolicy: {{.Values.clusterAutoscaler.image.pullPolicy | default .Values.defaults.imagePullPolicy }}
          name: main
          securityContext:
            allowPrivilegeEscalation: false
          livenessProbe:
            httpGet:
              path: /health-check
              port: 8085
            initialDelaySeconds: 15
            periodSeconds: 20
          readinessProbe:
            httpGet:
              path: /health-check
              port: 8085
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            limits:
              cpu: 500m
              memory: 500Mi
            requests:
              cpu: 200m
              memory: 200Mi
      serviceAccountName: {{.Release.Name}}-{{.Values.serviceAccount.nameSuffix}}
      terminationGracePeriodSeconds: 10
{{- end }}
