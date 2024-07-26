{{ $name := "nodepool-operator" }}

---
apiVersion: v1
kind: Secret
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
data:
  k3s_join_token: {{.Values.k3s.joinToken | b64enc}}
  k3s_server_public_host: {{.Values.k3s.serverPublicHost | b64enc}}
---

apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
  labels: &labels
    app: {{$name}}
    vector.dev/exclude: "true" # to exclude pods from being monitored by vector
spec:
  selector:
    matchLabels: *labels
  replicas: 1
  template:
    metadata:
      labels: *labels
    spec:
      securityContext:
        runAsNonRoot: true

      nodeSelector: {{.Values.nodepoolOperator.nodeSelector | toYaml | nindent 8}}
      tolerations: {{.Values.nodepoolOperator.tolerations |  toYaml | nindent 8}}

      affinity: 
        nodeAffinity: {{.Values.nodepoolOperator.nodeAffinity | toYaml | nindent 10 }}

      priorityClassName: {{ include "priority-class.name" .}}

      containers:
        - args:
            - --health-probe-bind-address=:8081
            - --metrics-bind-address=:9090
            - --leader-elect
          env:
            - name: MAX_CONCURRENT_RECONCILES
              value: "5"

            - name: ENABLE_NODEPOOLS
              value: "true"

            - name: KLOUDLITE_RELEASE
              value: {{.Values.kloudliteRelease}}

            - name: ACCOUNT_NAME
              value: {{.Values.accountName}}

            - name: CLUSTER_NAME
              value: {{.Values.clusterName}}

            - name: "K3S_JOIN_TOKEN"
              valueFrom:
                secretKeyRef:
                  name: {{$name}}
                  key: k3s_join_token

            - name: "K3S_SERVER_PUBLIC_HOST"
              valueFrom:
                secretKeyRef:
                  name: {{$name}}
                  key: k3s_server_public_host

            - name: "TF_STATE_SECRET_NAMESPACE"
              value: "{{.Release.Namespace}}"

            - name: "JOBS_NAMESPACE"
              value: "{{.Release.Namespace}}"

            - name: "IAC_JOB_IMAGE"
              value: {{.Values.nodepoolJob.image.repository}}:{{.Values.nodepoolJob.image.tag | default (.Values.kloudliteRelease | default .Chart.AppVersion)}}

          image: {{.Values.nodepoolOperator.image.repository}}:{{.Values.nodepoolOperator.image.tag | default (.Values.kloudliteRelease | default .Chart.AppVersion)}}
          imagePullPolicy:  {{.Values.nodepoolOperator.image.pullPolicy | default (.Values.imagePullPolicy | default "IfNotPresent")}}
          name: manager
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL

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
          resources: {{.Values.nodepoolOperator.resources | toYaml | nindent 12}}

      serviceAccountName: {{ include "service-account.name" .}}
      terminationGracePeriodSeconds: 10
