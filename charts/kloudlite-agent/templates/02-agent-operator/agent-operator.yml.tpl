{{ $name := .Values.agentOperator.name }}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
  labels: &labels
    app: {{$name}}
    control-plane: {{$name}}
    vector.dev/exclude: "true" # to exclude pods from being monitored by vector
  annotations:
    checksum/cluster-identity-secret: {{ include (print $.Template.BasePath "/secrets/cluster-identity-secret.yml.tpl") . | sha256sum }}
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
      nodeSelector: {{.Values.agentOperator.nodeSelector | default .Values.nodeSelector | toYaml | nindent 8}}
      tolerations: {{.Values.agentOperator.tolerations | default .Values.tolerations | toYaml | nindent 8}}

      affinity:
        nodeAffinity: {{.Values.agentOperator.nodeAffinity | default dict | toYaml | nindent 10}}

      priorityClassName: kloudlite-critical

      containers:
        - args:
            - --health-probe-bind-address=:8081
            - --metrics-bind-address=:9090
            - --leader-elect
          env:
            - name: MAX_CONCURRENT_RECONCILES
              value: "5"

            - name: CLUSTER_INTERNAL_DNS
              value: {{.Values.clusterInternalDNS}}

            - name: KLOUDLITE_DNS_SUFFIX
              value: {{.Values.kloudliteDNSSuffix}}

            {{ include "environment-operator-env" . | nindent 12 }}
            {{ include "resource-watcher-env" . | nindent 12 }}
            {{ include "router-operator-env" . | nindent 12 }}
            {{ include "msvc-mongo-operator-env" . | nindent 12 }}
            {{ include "helmchart-operator-env" . | nindent 12 }}

          {{- $imageTag := .Values.agentOperator.image.tag | default (include "image-tag" .) }}
          image: {{.Values.agentOperator.image.repository}}:{{$imageTag}}
          imagePullPolicy: {{ .Values.agentOperator.image.pullPolicy | default (include "image-pull-policy" $imageTag) }}
          name: manager
          securityContext:
            allowPrivilegeEscalation: false
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8081
            initialDelaySeconds: 15
            periodSeconds: 20
          readinessProbe:
            httpGet:
              path: /readyz
              port: 8081
            initialDelaySeconds: 5
            periodSeconds: 10
          resources: {{.Values.agentOperator.resources | toYaml | nindent 12}}

      serviceAccountName: {{include "serviceAccountName" .}}
      terminationGracePeriodSeconds: 10

---

apiVersion: v1
kind: Service
metadata:
  name: {{$name}}-metrics
  namespace: {{.Release.Namespace}}
  labels: &labels
    app: {{$name}}
spec:
  ports:
    - name: metrics
      port: 9090
      protocol: TCP
  selector: *labels
