{{if .Values.operators.byocOperator.enabled}}
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Values.operators.byocOperator.name}}
  namespace: {{.Release.Namespace}}
  labels:
    control-plane: {{.Values.operators.byocOperator.name}}
spec:
  ports:
    - name: https
      port: 8443
      protocol: TCP
      targetPort: https
  selector:
    control-plane: {{.Values.operators.byocOperator.name}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: {{.Values.operators.byocOperator.name}}
  name: {{.Values.operators.byocOperator.name}}
  namespace: {{.Release.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: {{.Values.operators.byocOperator.name}}
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: manager
      labels:
        control-plane: {{.Values.operators.byocOperator.name}}
    spec:
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
            - /manager
          image: {{.Values.operators.byocOperator.image}}
          imagePullPolicy: {{.Values.operators.byocOperator.ImagePullPolicy | default .Values.imagePullPolicy }}
          env:
            - name: RECONCILE_PERIOD
              value: 30s

            - name: MAX_CONCURRENT_RECONCILES
              value: "1"

            - name: KAFKA_TOPIC_NAMESPACE
              value: {{.Release.Namespace}}

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
      securityContext:
        runAsNonRoot: true
      serviceAccountName: "{{.Values.clusterSvcAccount}}"
      terminationGracePeriodSeconds: 10
{{end}}
