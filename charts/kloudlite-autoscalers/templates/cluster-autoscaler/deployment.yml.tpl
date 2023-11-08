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

      {{- if .Values.tolerations }}
      tolerations: {{.Values.tolerations | toYaml | nindent 8}}
      {{- end }}

      {{- if .Values.nodeSelector}}
      nodeSelector: {{.Values.nodeSelector | toYaml | nindent 8}}
      {{- end }}

      {{- if .Values.preferOperatorsOnMasterNodes }}
      affinity:
        nodeAffinity: {{include "preferred-node-affinity-to-masters" . | nindent 10 }}
      {{- end }}

      containers:
        - command:
            - /cluster-autoscaler
          args:
            - --cloud-provider=kloudlite
            - --logtostderr=true
            - --stderrthreshold=info
            - scale-down-unneeded-time=1m
          image: {{.Values.clusterAutoscaler.image.repository}}:{{.Values.clusterAutoscaler.image.tag | default .Values.defaults.imageTag }}
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
      serviceAccountName: {{.Values.svcAccountName}}
      terminationGracePeriodSeconds: 10
{{- end }}
