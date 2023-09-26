{{- if .Values.operators.helmChartsOperator.enabled }}
{{ $name := .Values.operators.helmChartsOperator.name }}

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{$name}}
  namespace: {{.Release.Namespace}}
  labels: &labels
    app: {{$name}}
    control-plane: {{$name}}
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
      {{- if .Values.operators.helmChartsOperator.configuration.affinity }}
      affinity: {{.Values.operators.helmChartsOperator.configuration.affinity | toYaml | nindent 8}}
      {{- end }}

      {{ include "node-selector-and-tolerations" . | nindent 6 }}

      initContainers:
        - name: init-container
          image: busybox:latest
          command: ['sh', '-c', 'chown 1717:1717 -R /tmp/helm-repository-cache']
          securityContext:
            allowPrivilegeEscalation: true
            runAsUser: 0
          volumeMounts:
            - name: repository-cache
              mountPath: /tmp/helm-repository-cache

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

          image: {{.Values.operators.helmChartsOperator.image}}
          imagePullPolicy: {{.Values.operators.helmChartsOperator.imagePullPolicy | default .Values.imagePullPolicy }}

          env:
            - name: RECONCILE_PERIOD
              value: "30s"
            - name: MAX_CONCURRENT_RECONCILES
              value: "1"
            - name: HELM_REPOSITORY_CACHE_DIR
              value: "/tmp/helm-repository-cache"
          volumeMounts:
            - name: repository-cache
              mountPath: /tmp/helm-repository-cache

          name: manager
          securityContext:
            runAsNonRoot: true
            runAsUser: 1717
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
              cpu: 150m
              memory: 200Mi
            requests:
              cpu: 100m
              memory: 150Mi
      volumes:
        - name:  repository-cache
          hostPath:
            path: /mnt/helm-charts-repo-cache
            type: DirectoryOrCreate
      serviceAccountName: {{.Values.svcAccountName}}
      terminationGracePeriodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  labels: &labels
    app: {{$name}}
    control-plane: {{$name}}
  name: {{$name}}-metrics-service
  namespace: {{.Release.Namespace}}
spec:
  ports:
    - name: https
      port: 8443
      protocol: TCP
      targetPort: https
  selector: *labels

{{- end }}
