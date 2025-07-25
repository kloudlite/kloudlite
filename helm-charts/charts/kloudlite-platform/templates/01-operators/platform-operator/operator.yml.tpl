---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: kloudlite-platform-operator
  name: kloudlite-platform-operator
  namespace: {{.Release.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: kloudlite-platform-operator
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: manager
      labels:
        control-plane: kloudlite-platform-operator
    spec:
      affinity: {{ .Values.operators.platformOperator.affinity | toYaml | nindent 8 }}
      tolerations: {{ (.Values.operators.platformOperator.tolerations | default .Values.scheduling.stateless.tolerations) | toYaml | nindent 8 }}
      nodeSelector: {{ (.Values.operators.platformOperator.nodeSelector | default .Values.scheduling.stateless.nodeSelector) | toYaml | nindent 8 }}

      containers:
        - args:
            - --secure-listen-address=0.0.0.0:8443
            - --upstream=http://127.0.0.1:8080/
            - --logtostderr=true
            - --v=0
          image: gcr.io/kubebuilder/kube-rbac-proxy:v0.13.0
          name: kube-rbac-proxy
          ports:
            - containerPort: 8443
              name: https
              protocol: TCP
          resources:
            limits:
              cpu: 20m
              memory: 30Mi
            requests:
              cpu: 5m
              memory: 10Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
        - args:
            - --health-probe-bind-address=:8081
            - --metrics-bind-address=127.0.0.1:8080
            - --leader-elect
          image: {{.Values.operators.platformOperator.image.repository}}:{{.Values.operators.platformOperator.image.tag | default (include "image-tag" .) }}
          imagePullPolicy: {{.Values.operators.platformOperator.image.pullPolicy | default (include "image-pull-policy" .)}}
          env:
            - name: MAX_CONCURRENT_RECONCILES
              value: "5"

            - name: CLUSTER_INTERNAL_DNS
              value: "{{.Values.clusterInternalDNS}}"

            {{ include "router-operator-env" . | nindent 12 }}
            {{ include "helmchart-operator-env" . | nindent 12 }}

            {{- /* platform operator should not need it, though */}}
            - name: KLOUDLITE_DNS_SUFFIX
              value: ""

            - name: KLOUDLITE_NAMESPACE
              value: "{{.Release.Namespace}}"

          livenessProbe:
            httpGet:
              path: /healthz
              port: 8081
            initialDelaySeconds: 15
            periodSeconds: 20
          name: manager
          readinessProbe:
            httpGet:
              path: /readyz
              port: 8081
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            limits:
              cpu: 200m
              memory: 200Mi
            requests:
              cpu: 150m
              memory: 150Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
      securityContext:
        runAsNonRoot: true
      serviceAccountName: {{.Values.serviceAccounts.clusterAdmin.name}}
      terminationGracePeriodSeconds: 10
