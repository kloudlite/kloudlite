{{if .Values.operators.byocClientOperator.enabled}}
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Values.operators.byocClientOperator.name}}
  namespace: {{.Release.Namespace}}
  labels:
    control-plane: {{.Values.operators.byocClientOperator.name}}
spec:
  ports:
    - name: https
      port: 8443
      protocol: TCP
      targetPort: https
  selector:
    control-plane: {{.Values.operators.byocClientOperator.name}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: {{.Values.operators.byocClientOperator.name}}
  name: {{.Values.operators.byocClientOperator.name}}
  namespace: {{.Release.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: {{.Values.operators.byocClientOperator.name}}
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: manager
      labels:
        control-plane: {{.Values.operators.byocClientOperator.name}}
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
          image: {{.Values.operators.byocClientOperator.image}}
          imagePullPolicy: {{.Values.operators.byocClientOperator.ImagePullPolicy | default .Values.imagePullPolicy }}
          env:
            - name: RECONCILE_PERIOD
              value: 30s

            - name: MAX_CONCURRENT_RECONCILES
              value: "1"

            - name: HELM_RELEASE_NAMESPACE
              value: {{.Release.Namespace}}

            - name: GRPC_ADDR
              value: {{.Values.messageOfficeApi.grpcAddr}}

            - name: ACCOUNT_NAME
              value: {{.Values.accountName }}

            - name: CLUSTER_TOKEN
              valueFrom:
                secretKeyRef:
                  name: {{.Values.clusterIdentitySecretName}}
                  key: CLUSTER_TOKEN
                  optional: true

            - name: ACCESS_TOKEN
              valueFrom:
                secretKeyRef:
                  name: {{.Values.clusterIdentitySecretName}}
                  key: ACCESS_TOKEN
                  optional: true

            {{/* - name: CLUSTER_NAME */}}
            {{/*   value: {{.Values.clusterName}} */}}
            {{/**/}}
            {{/* - name: RELEASE_NAME */}}
            {{/*   value: {{.Release.Name}} */}}

            {{/* - name: KAFKA_TOPIC_BYOC_CLIENT_UPDATES */}}
            {{/*   value: {{.Values.kafka.topicBYOCClientUpdates}} */}}
            {{/**/}}
            {{/* - name: KAFKA_BROKERS */}}
            {{/*   valueFrom: */}}
            {{/*     secretKeyRef: */}}
            {{/*       name: {{.Values.kafka.secretName}} */}}
            {{/*       key: KAFKA_BROKERS */}}
            {{/**/}}
            {{/* - name: KAFKA_SASL_USERNAME */}}
            {{/*   valueFrom: */}}
            {{/*     secretKeyRef: */}}
            {{/*       name: {{.Values.kafka.secretName}} */}}
            {{/*       key: KAFKA_SASL_USERNAME */}}
            {{/**/}}
            {{/* - name: KAFKA_SASL_PASSWORD */}}
            {{/*   valueFrom: */}}
            {{/*     secretKeyRef: */}}
            {{/*       name: {{.Values.kafka.secretName}} */}}
            {{/*       key: KAFKA_SASL_PASSWORD */}}
            
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
      serviceAccountName: "{{.Values.svcAccountName}}"
      terminationGracePeriodSeconds: 10
{{end}}
