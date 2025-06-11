---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: {{.Release.Name}}
  name: {{.Release.Name}}
  namespace: {{.Release.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: {{.Release.Name}}
  template:
    metadata:
      labels:
        control-plane: {{.Release.Name}}
    spec:
      affinity: {{ .Values.affinity | toJson }}
      tolerations: {{ .Values.tolerations | toJson }}
      nodeSelector: {{ .Values.nodeSelector | toJson }}

      containers:
        - args:
            - --health-probe-bind-address=:8081
            - --metrics-bind-address=127.0.0.1:8080
            - --leader-elect
          image: "{{.Values.image.name}}:{{ .Values.image.tag | default .Chart.AppVersion }}"
          imagePullPolicy: {{ .Values.pullPolicy | default "IfNotPresent" }}
          env:
            - name: MAX_CONCURRENT_RECONCILES
              value: "5"

            - name: CLUSTER_INTERNAL_DNS
              value: "{{.Values.clusterInternalDNS}}"

            {{- /* platform operator should not need it, though */}}
            - name: KLOUDLITE_DNS_SUFFIX
              value: ""

            - name: KLOUDLITE_NAMESPACE
              valueFrom:
                fieldRef: 
                  fieldPath: "metadata.namespace"

            - name: DEFAULT_CLUSTER_ISSUER
              value: {{.Values.certManager.clusterIssuer}}

            - name: DEFAULT_INGRESS_CLASS
              {{- /* value: {{.Values.operators.platformOperator.configuration.ingressClassName}} */}}
              value: {{.Values.ingress.ingressClass}}

            - name: CERTIFICATE_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: "metadata.namespace"

            - name: HELM_JOB_RUNNER_IMAGE
              value: "{{.Values.helmJobRunnerImage.name}}:{{.Values.helmJobRunnerImage.tag | default .Chart.AppVersion}}"

            - name: IAC_JOBS_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: "metadata.namespace"

            - name: IAC_JOB_IMAGE
              value: "{{.Values.infrastructureAsCodeImage.name}}:{{.Values.infrastructureAsCodeImage.tag | default .Chart.AppVersion}}"

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
              cpu: 120m
              memory: 120Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
      securityContext:
        runAsNonRoot: true
      serviceAccountName: {{ include "service-account.name" . }}
      terminationGracePeriodSeconds: 10
