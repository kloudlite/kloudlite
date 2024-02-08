{{- if .Values.operators.helmOperator.enabled }}

{{- $operatorName := .Values.operators.helmOperator.name }}
---
apiVersion: v1
data:
  controller_manager_config.yaml: "apiVersion: controller-runtime.sigs.k8s.io/v1alpha1\n\
    kind: ControllerManagerConfig\nhealth:\n  healthProbeBindAddress: :8081\nmetrics:\n\
    \  bindAddress: 127.0.0.1:8080\nleaderElection:\n  leaderElect: true\n  resourceName:\
    \ 811c9dc5.kloudlite.io\n"
kind: ConfigMap
metadata:
  name: {{$operatorName}}-manager-config
  namespace: {{.Release.Namespace}}
---
apiVersion: v1
kind: Service
metadata:
  labels:
    control-plane: {{$operatorName}}
  name: {{$operatorName}}
  namespace: {{.Release.Namespace}}
spec:
  ports:
    - name: https
      port: 8443
      protocol: TCP
      targetPort: https
  selector:
    control-plane: {{$operatorName}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: {{$operatorName}}
  name: {{$operatorName}}
  namespace: {{.Release.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: {{$operatorName}}
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: manager
      labels:
        control-plane: {{$operatorName}}
    spec:
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
            requests:
              cpu: 10m
              memory: 64Mi
        - args:
            - --health-probe-bind-address=:8081
            - --metrics-bind-address=127.0.0.1:8080
            - --leader-elect
            - --leader-election-id=helm
          image: {{.Values.operators.helmOperator.image}}
          imagePullPolicy: {{.Values.operators.helmOperator.ImagePullPolicy | default .Values.imagePullPolicy }}
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
          securityContext:
            allowPrivilegeEscalation: false
          resources:
            requests:
              cpu: 120m
              memory: 240Mi
      securityContext:
        runAsNonRoot: true
      serviceAccountName: {{.Values.svcAccountName}}
      terminationGracePeriodSeconds: 10
      {{- if .Tolerations }}
      tolerations: {{ .Tolerations | toYaml | nindent 8 }}
      {{- end }}

      {{- if .NodeSelector }}
      nodeSelector: {{ .NodeSelector | toYaml | nindent 8 }}
      {{- end }}
{{- end}}
