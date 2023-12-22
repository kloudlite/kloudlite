{{if .Values.operators.wgOperator.enabled}}
{{ $name := .Values.operators.wgOperator.name }}
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/component: kube-rbac-proxy
    app.kubernetes.io/name: service
    app.kubernetes.io/part-of: {{$name}}
    control-plane: {{$name}}
  name: {{$name}}
  namespace: {{.Release.Namespace}}
spec:
  ports:
    - name: https
      port: 8443
      protocol: TCP
      targetPort: https
  selector:
    control-plane: {{$name}}

---

apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/component: manager
    app.kubernetes.io/name: deployment
    app.kubernetes.io/part-of: {{$name}}
    control-plane: {{$name}}
  name: {{$name}}
  namespace: {{.Release.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels: &labels
      control-plane: {{$name}}
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: manager
      labels: *labels
    spec:
      nodeSelector: {{.Values.operators.wgOperator.nodeSelector | default .Values.defaults.nodeSelector | toYaml | nindent 8}}
      tolerations: {{.Values.operators.wgOperator.tolerations | default .Values.defaults.tolerations | toYaml | nindent 8}}
      affinity:
        nodeAffinity:
          {{ if .Values.preferOperatorsOnMasterNodes }}
          {{ include "preferred-node-affinity-to-masters" . | nindent 10 }}
          {{ end }}
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: kubernetes.io/arch
                    operator: In
                    values:
                      - amd64
                      - arm64
                      - ppc64le
                      - s390x
                  - key: kubernetes.io/os
                    operator: In
                    values:
                      - linux

      priorityClassName: kloudlite-critical
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
              cpu: 30m
              memory: 30Mi
            requests:
              cpu: 20m
              memory: 20Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
        - args:
            - --health-probe-bind-address=:8081
            - --metrics-bind-address=127.0.0.1:8080
            - --leader-elect

          env:
            - name: RECONCILE_PERIOD
              value: "30s"
              
            - name: MAX_CONCURRENT_RECONCILES
              value: "5"

            - name: WG_POD_CIDR
              value: {{.Values.operators.wgOperator.configuration.podCIDR}}

            - name: WG_SVC_CIDR
              value: {{.Values.operators.wgOperator.configuration.svcCIDR}}

            - name: DNS_HOSTED_ZONE
              value: {{.Values.operators.wgOperator.configuration.dnsHostedZone}}

            - name: CLUSTER_INTERNAL_DNS
              value: {{.Values.clusterInternalDNS}}

          image: {{.Values.operators.wgOperator.image.repository}}:{{.Values.operators.wgOperator.image.tag | default .Values.defaults.imageTag | default .Chart.AppVersion}}
          imagePullPolicy: {{.Values.operators.wgOperator.image.pullPolicy | default .Values.imagePullPolicy}}
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
              cpu: 100m
              memory: 100Mi
            requests:
              cpu: 100m
              memory: 100Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
      securityContext:
        runAsNonRoot: true
      serviceAccountName: {{ include "serviceAccountName" . | squote}}
      terminationGracePeriodSeconds: 10
{{end}}
