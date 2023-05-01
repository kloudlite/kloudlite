{{if .Values.operators.clusterOperator.enabled}}
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Values.operators.clusterOperator.name}}
  namespace: {{.Release.Namespace}}
  labels:
    control-plane: {{.Values.operators.clusterOperator.name}}
spec:
  ports:
    - name: https
      port: 8443
      protocol: TCP
      targetPort: https
  selector:
    control-plane: {{.Values.operators.clusterOperator.name}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: {{.Values.operators.clusterOperator.name}}
  name: {{.Values.operators.clusterOperator.name}}
  namespace: {{.Release.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: {{.Values.operators.clusterOperator.name}}
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: manager
      labels:
        control-plane: {{.Values.operators.clusterOperator.name}}
    spec:
      affinity:
        nodeAffinity:
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
          image: {{.Values.operators.clusterOperator.image}}
          imagePullPolicy: {{.Values.operators.clusterOperator.ImagePullPolicy | default .Values.imagePullPolicy }}
          
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
              cpu: 40m
              memory: 80Mi
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
