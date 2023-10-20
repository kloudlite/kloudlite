{{- $clusterOperator := .Values.operators.clusterOperator }} 

{{if $clusterOperator.enabled}}
---
apiVersion: v1
kind: Service
metadata:
  labels:
    app.kubernetes.io/component: kube-rbac-proxy
    app.kubernetes.io/name: service
    app.kubernetes.io/part-of: {{$clusterOperator.name}}
    control-plane: {{$clusterOperator.name}}
  name: {{$clusterOperator.name}}
  namespace: {{.Release.Namespace}}
spec:
  ports:
    - name: https
      port: 8443
      protocol: TCP
      targetPort: https
  selector:
    control-plane: {{$clusterOperator.name}}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{$clusterOperator.name}}-cloudflare-params
  namespace: {{.Release.Namespace}}
data:
  api_token: {{$clusterOperator.configuration.cloudflare.apiToken | b64enc | quote }}
  base_domain: {{$clusterOperator.configuration.cloudflare.baseDomain | b64enc | quote }}
  zone_id: {{$clusterOperator.configuration.cloudflare.zoneId | b64enc | quote }}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/component: manager
    app.kubernetes.io/name: deployment
    app.kubernetes.io/part-of: {{$clusterOperator.name}}
    control-plane: {{$clusterOperator.name}}
  name: {{$clusterOperator.name}}
  namespace: {{.Release.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels: &labels
      control-plane: {{$clusterOperator.name}}
  template:
    metadata:
      annotations:
        kubectl.kubernetes.io/default-container: manager
      labels: *labels
    spec:
      affinity:
        {{- if .Values.preferOperatorsOnMasterNodes }}
        {{include "preferred-node-affinity-to-masters" . | nindent 10 }}
        {{- end }}
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
          command:
            - /manager
          
          env:
            - name: RECONCILE_PERIOD
              value: "30s"
              
            - name: MAX_CONCURRENT_RECONCILES
              value: "5"

            - name: CLOUDFLARE_API_TOKEN
              valueFrom:
                secretKeyRef:
                  name: {{$clusterOperator.name}}-cloudflare-params
                  key: api_token

            - name: CLOUDFLARE_ZONE_ID
              valueFrom:
                secretKeyRef:
                  name: {{$clusterOperator.name}}-cloudflare-params
                  key: zone_id

            - name: CLOUDFLARE_DOMAIN
              valueFrom:
                secretKeyRef:
                  name: {{$clusterOperator.name}}-cloudflare-params
                  key: base_domain

            - name: KL_S3_BUCKET_NAME
              {{include "is-required" 
                  (list
                    $clusterOperator.configuration.IACStateStore.s3BucketName 
                    ".operators.clusterOperator.configuration.IACStateStore.s3BucketName is required" 
                  )
              }}
              value: {{$clusterOperator.configuration.IACStateStore.s3BucketName}}

            - name: KL_S3_BUCKET_REGION
            {{- if not $clusterOperator.configuration.IACStateStore.s3BucketRegion}}}
            {{ fail ".operators.clusterOperator.configuration.IACStateStore.s3BucketRegion is required" }}
            {{- end }}
              value: {{.Values.operators.clusterOperator.configuration.IACStateStore.s3BucketRegion}}

            - name: MESSAGE_OFFICE_GRPC_ADDR
              value: "{{.Values.routers.messageOfficeApi.name}}.{{.Values.baseDomain}}:443"

            - name: KL_AWS_ACCESS_KEY
            {{- if not $clusterOperator.configuration.IACStateStore.accessKey}}}
            {{ fail ".operators.clusterOperator.configuration.IACStateStore.accessKey is required" }}
            {{- end }}
              value: "{{$clusterOperator.configuration.IACStateStore.accessKey}}"

            - name: KL_AWS_SECRET_KEY
            {{- if not $clusterOperator.configuration.IACStateStore.secretKey}}}
            {{ fail ".operators.clusterOperator.configuration.IACStateStore.secretKey is required" }}
            {{- end }}
              value: "{{ $clusterOperator.configuration.IACStateStore.secretKey }}"


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
              cpu: 50m
              memory: 50Mi
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
      serviceAccountName: {{.Values.clusterSvcAccount | squote}}
      terminationGracePeriodSeconds: 10
{{end}}

