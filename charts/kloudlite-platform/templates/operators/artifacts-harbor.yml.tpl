{{ if .Values.operators.artifactsHarbor.enabled }}
{{ $name := .Values.operators.artifactsHarbor.name}} 

{{ $harborSecretName := printf "harbor-%s-admin-creds" .Release.Name }} 

---

apiVersion: v1
kind: Secret
metadata:
  name: {{$harborSecretName}}
  namespace: {{.Release.Namespace}}
stringData:
  API_VERSION: {{.Values.harbor.apiVersion}}
  ADMIN_USERNAME: {{.Values.harbor.adminUsername}}
  ADMIN_PASSWORD: {{.Values.harbor.adminPassword}}
  IMAGE_REGISTRY_HOST: {{.Values.harbor.imageRegistryHost}}
  WEBHOOK_ENDPOINT: {{.Values.harbor.webhookEndpoint}}
  WEBHOOK_NAME: {{.Values.harbor.webhookName}}
  WEBHOOK_AUTHZ: {{.Values.harbor.webhookAuthz}}
  {{/* SERVICE_ACCOUNT: {{.Values.harbor.serviceAccount}} */}}

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
      securityContext:
        runAsNonRoot: true

      {{- if .Values.nodeSelector}}
      nodeSelector: {{.Values.nodeSelector | toYaml | nindent 8}}
      {{- end }}

      {{- if .Values.tolerations }}
      tolerations: {{.Values.tolerations | toYaml | nindent 8}}
      {{- end }}
      
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
          env:
            - name: RECONCILE_PERIOD
              value: "30s"

            - name: MAX_CONCURRENT_RECONCILES
              value: "5"

            - name: HARBOR_API_VERSION
              valueFrom:
                secretKeyRef:
                  name: {{$harborSecretName}}
                  key: API_VERSION

            - name: HARBOR_ADMIN_USERNAME
              valueFrom:
                secretKeyRef:
                  name: {{$harborSecretName}}
                  key: ADMIN_USERNAME

            - name: HARBOR_ADMIN_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{$harborSecretName}}
                  key: ADMIN_PASSWORD

            - name: HARBOR_IMAGE_REGISTRY_HOST
              valueFrom:
                secretKeyRef:
                  name: {{$harborSecretName}}
                  key: IMAGE_REGISTRY_HOST

            - name: HARBOR_WEBHOOK_AUTHZ
              valueFrom:
                secretKeyRef:
                  name: {{$harborSecretName}}
                  key: WEBHOOK_AUTHZ

            - name: HARBOR_WEBHOOK_ENDPOINT
              valueFrom:
                secretKeyRef:
                  name: {{$harborSecretName}}
                  key: WEBHOOK_ENDPOINT

            - name: HARBOR_WEBHOOK_NAME
              valueFrom:
                secretKeyRef:
                  name: {{$harborSecretName}}
                  key: WEBHOOK_NAME

{{/*            - name: SERVICE_ACCOUNT_NAME*/}}
{{/*              valueFrom:*/}}
{{/*                secretKeyRef:*/}}
{{/*                  name: harbor-admin-creds*/}}
{{/*                  key: SERVICE_ACCOUNT*/}}
          
          image: {{.Values.operators.artifactsHarbor.image}}
          imagePullPolicy: {{.Values.operators.artifactsHarbor.ImagePullPolicy | default .Values.imagePullPolicy }}
          name: manager
          securityContext:
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
              cpu: 40m
              memory: 40Mi
            requests:
              cpu: 20m
              memory: 20Mi
      serviceAccountName: {{.Values.clusterSvcAccount}}
      terminationGracePeriodSeconds: 10

---

apiVersion: v1
kind: Service
metadata:
  name: {{$name}}-metrics-service
  namespace: {{.Release.Namespace}}
  labels: &labels
    app: {{$name}}
    control-plane: {{$name}}
spec:
  ports:
    - name: https
      port: 8443
      protocol: TCP
      targetPort: https
  selector: *labels
{{end}}
