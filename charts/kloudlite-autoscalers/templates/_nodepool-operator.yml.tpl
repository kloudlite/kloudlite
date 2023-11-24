{{if .Values.nodepools.enabled}}

{{- $name := "kl-nodepool-operator" }} 
{{- $providerCreds := "cloudprovider-credentials" }} 
---
apiVersion: v1
kind: Secret
metadata:
  name: {{$providerCreds}}
  namespace: {{.Release.Namespace}}
data:
  accessKey: "{{ .Values.cloudprovider.accessKey | b64enc }}"
  secretKey: "{{ .Values.cloudprovider.secretKey | b64enc }}"
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
            
            - name: CLOUD_PROVIDER_NAME
              value: {{.Values.cloudprovider.name}}
                  
            - name: CLOUD_PROVIDER_REGION
              value: {{.Values.cloudprovider.region}}

            - name: CLOUD_PROVIDER_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: {{$providerCreds}}
                  key: accessKey
                  optional: true

            - name: CLOUD_PROVIDER_SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: {{$providerCreds}}
                  key: secretKey
                  optional: true

            - name: K3S_JOIN_TOKEN
              value: {{.Values.k3sMasters.joinToken}}

            - name: K3S_SERVER_PUBLIC_HOST
              value: {{.Values.k3sMasters.publicHost}}

            - name: IAC_STATE_S3_BUCKET_NAME
              value: {{.Values.IACStateStore.bucketName}}

            - name: IAC_STATE_S3_BUCKET_REGION
              value: {{.Values.IACStateStore.bucketRegion}}

            - name: IAC_STATE_S3_BUCKET_DIR
              value: {{.Values.IACStateStore.bucketDir}}

            - name: KLOUDLITE_ACCOUNT_NAME
              value: {{.Values.accountName}}

            - name: KLOUDLITE_CLUSTER_NAME
              value: "{{.Values.clusterName}}"

          image: {{.Values.nodepools.image.repository}}:{{.Values.nodepools.image.tag | default .Values.defaults.imageTag}}
          imagePullPolicy: {{.Values.nodepools.image.pullPolicy | default .Values.defaults.imagePullPolicy}}
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
      serviceAccountName: {{include "service-account-name" .}}
      terminationGracePeriodSeconds: 10
{{end}}

