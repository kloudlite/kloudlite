{{ if .Values.operators.agentOperator.enabled }}
{{ $name := .Values.operators.agentOperator.name }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
  labels: &labels
    app: {{$name}}
    control-plane: {{$name}}
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
      nodeSelector: {{.Values.operators.agentOperator.nodeSelector | default .Values.defaults.nodeSelector | toYaml | nindent 8}}
      tolerations: {{.Values.operators.agentOperator.tolerations | default .Values.defaults.tolerations | toYaml | nindent 8}}

      {{- if .Values.preferOperatorsOnMasterNodes }}
      affinity:
        nodeAffinity:
          {{include "preferred-node-affinity-to-masters" . | nindent 12 }}
      {{- end }}

      priorityClassName: kloudlite-critical

      containers:
        - args:
            - --secure-listen-address=0.0.0.0:8443
            - --upstream=http://127.0.0.1:8080/
            - --logtostderr=true
            - --v=0
          image: gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
          name: kube-rbac-proxy
          ports:
            - containerPort: 8443
              name: https
              protocol: TCP
          resources:
            limits:
              cpu: 20m
              memory: 20Mi
            requests:
              cpu: 5m
              memory: 10Mi

        - command:
            - /manager
          args:
            - --health-probe-bind-address=:8081
            - --metrics-bind-address=127.0.0.1:8080
            - --leader-elect
          env:
            - name: RECONCILE_PERIOD
              value: "30s"

            - name: MAX_CONCURRENT_RECONCILES
              value: "5"

            {{- /* CLUSTER_INTERNAL_DNS="cluster.local" */}}
            {{- /* SVC_ACCOUNT_NAME="kloudlite-svc-account" */}}
            - name: CLUSTER_INTERNAL_DNS
              value: {{.Values.clusterInternalDNS}}

            {{- /* for: project operator */}}
            - name: SVC_ACCOUNT_NAME
              value: "kloudlite-svc-account"

            {{- /* for: resource watcher */}}
            - name: OPERATORS_NAMESPACE
              value: {{.Release.Namespace}}

            - name: GRPC_ADDR
              value: {{.Values.messageOfficeGRPCAddr}}

            - name: ACCOUNT_NAME
              value: {{.Values.accountName }}

            - name: CLUSTER_NAME
              value: {{.Values.clusterName }}

            - name: CLUSTER_IDENTITY_SECRET_NAME
              value: {{.Values.clusterIdentitySecretName}}

            - name: CLUSTER_IDENTITY_SECRET_NAMESPACE
              value: {{.Release.Namespace}}

            - name: ACCESS_TOKEN
              valueFrom:
                secretKeyRef:
                  name: {{.Values.clusterIdentitySecretName}}
                  key: ACCESS_TOKEN
                  optional: true

            {{- /* for: nodepool operator */}}
            - name: KLOUDLITE_ACCOUNT_NAME
              value: {{.Values.accountName}}
            - name: KLOUDLITE_CLUSTER_NAME
              value: {{.Values.clusterName}}

            - name: "K3S_JOIN_TOKEN"
              value: {{.Values.operators.agentOperator.configuration.k3sJoinToken}}
            - name: "K3S_SERVER_PUBLIC_HOST"
              value: {{.Values.operators.agentOperator.configuration.k3sServerPublicHost}}

            - name: CLOUD_PROVIDER_NAME
              value: {{.Values.operators.agentOperator.configuration.cloudProvider.name}}
            - name: CLOUD_PROVIDER_REGION
              value: {{.Values.operators.agentOperator.configuration.cloudProvider.region}}

            {{- /* for: routers */}}
            - name: ACME_EMAIL
              value: {{.Values.operators.agentOperator.configuration.letsEncryptSupportEmail}}
            - name: WORKSPACE_ROUTE_SWITCHER_SERVICE
              value: "env-route-switcher"
            - name: WORKSPACE_ROUTE_SWITCHER_PORT
              value: "80"

          image: {{.Values.operators.agentOperator.image.repository}}:{{.Values.operators.agentOperator.image.tag | default .Values.defaults.imageTag | default .Chart.AppVersion}}
          imagePullPolicy: {{.Values.operators.agentOperator.image.pullPolicy | default .Values.imagePullPolicy}}
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
          resources:
            limits:
              cpu: 100m
              memory: 100Mi
            requests:
              cpu: 20m
              memory: 20Mi
      serviceAccountName: {{include "serviceAccountName" .}}
      terminationGracePeriodSeconds: 10

---

apiVersion: v1
kind: Service
metadata:
  name: {{$name}}-metrics-service
  namespace: {{.Release.Namespace}}
  labels: &labels
    app: {{$name}}
    control-plane: {{$name}}
spec:
  ports:
    - name: https
      port: 8443
      protocol: TCP
      targetPort: https
  selector: *labels
{{end}}
