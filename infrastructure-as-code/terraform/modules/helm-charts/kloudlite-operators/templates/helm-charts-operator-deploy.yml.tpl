apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${name}
  namespace: ${namespace}
  annotations:
    kloudlite.io/description: |+
      Helm Charts Operator, by kloudlite labs, is a Kubernetes operator for managing Helm charts.
spec:
  selector:
    matchLabels:
      app: ${name}
  replicas: 1
  template:
    metadata:
      labels:
        app: ${name}
      annotations:
        kubectl.kubernetes.io/default-container: manager
    spec:
      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - preference:
                matchExpressions:
                  - key: node-role.kubernetes.io/master
                    operator: In
                    values:
                      - "true"
              weight: 1

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

          image: ${image}
          imagePullPolicy: ${image_pull_policy}

          env:
            - name: RECONCILE_PERIOD
              value: "30s"
            - name: MAX_CONCURRENT_RECONCILES
              value: "1"
            - name: RUNNING_IN_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
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
      serviceAccountName: ${svc_account_name}
      terminationGracePeriodSeconds: 10
