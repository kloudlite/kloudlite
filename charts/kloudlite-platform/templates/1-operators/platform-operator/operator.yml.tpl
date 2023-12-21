{{if .Values.operators.platformOperator.enabled}}
---
apiVersion: v1
kind: Service
metadata:
  name: kloudlite-platform-operator
  namespace: {{.Release.Namespace}}
  labels:
    control-plane: kloudlite-platform-operator
spec:
  ports:
    - name: https
      port: 8443
      protocol: TCP
      targetPort: https
  selector:
    control-plane: kloudlite-platform-operator
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
      {{- if .Values.operators.preferOperatorsOnMasterNodes }}
      affinity:
        nodeAffinity: {{include "preferred-node-affinity-to-masters" . | nindent 10 }}
      {{- end }}
      tolerations: {{.Values.operators.platformOperator.configuration.tolerations | toYaml | nindent 8 }}
      nodeSelector: {{.Values.operators.platformOperator.configuration.nodeSelector | toYaml | nindent 8 }}
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
          command:
            - /app/app
          image: {{.Values.operators.platformOperator.image}}
          imagePullPolicy: {{.Values.global.imagePullPolicy }}
          env:
            - name: RECONCILE_PERIOD
              value: 30s
            - name: SVC_ACCOUNT_NAME
              value: "kloudlite-svc-account"

            - name: MAX_CONCURRENT_RECONCILES
              value: "5"
            {{ include "project-operator-env" . | nindent 12 }}
            {{ include "cluster-operator-env" . | nindent 12 }}
            {{ include "router-operator-env" . | nindent 12 }}


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
      serviceAccountName: {{.Values.global.clusterSvcAccount}}
      terminationGracePeriodSeconds: 10
{{end}}
