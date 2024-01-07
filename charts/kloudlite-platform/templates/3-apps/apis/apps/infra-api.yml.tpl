apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: infra-api
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.clusterSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 4 }}

  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.infraApi.image}}
      imagePullPolicy: {{.Values.global.imagePullPolicy }}
      {{if .Values.global.isDev}}
      args:
       - --dev
      {{end}}
      
      resourceCpu:
        min: "50m"
        max: "100m"
      resourceMemory:
        min: "50Mi"
        max: "100Mi"

      env:
        - key: ACCOUNTS_GRPC_ADDR
          value: "accounts-api:3001"

        - key: MONGO_DB_URI
          type: secret
          refName: mres-infra-db-creds
          refKey: URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-infra-db-creds
          refKey: DB_NAME

        - key: HTTP_PORT
          value: "3000"

        - key: GRPC_PORT
          value: "3001"

        - key: COOKIE_DOMAIN
          value: "{{.Values.global.cookieDomain}}"

        - key: NATS_URL
          value: "nats://nats:4222"

        - key: ACCOUNT_COOKIE_NAME
          value: kloudlite-account

        - key: PROVIDER_SECRET_NAMESPACE
          value: {{.Release.Namespace}}

        - key: IAM_GRPC_ADDR
          value: "iam:3001"

        - key: NATS_STREAM
          value: {{.Values.envVars.nats.streams.resourceSync.name}}

        - key: SESSION_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.sessionKVBucket.name}}

        - key: MESSAGE_OFFICE_INTERNAL_GRPC_ADDR
          value: "message-office:3002"

        - key: AWS_ACCESS_KEY
          value: {{.Values.aws.accessKey}}

        - key: AWS_SECRET_KEY
          value: {{.Values.aws.secretKey}}

        - key: AWS_CF_STACK_S3_URL
          value: {{.Values.aws.cloudformation.stackS3URL}}

        - key: AWS_CF_PARAM_TRUSTED_ARN
          value: {{.Values.aws.cloudformation.params.trustedARN}}
        
        - key: AWS_CF_STACK_NAME_PREFIX
          value: {{.Values.aws.cloudformation.stackNamePrefix}}

        - key: AWS_CF_ROLE_NAME_PREFIX
          value: {{.Values.aws.cloudformation.roleNamePrefix}}

        - key: AWS_CF_INSTANCE_PROFILE_NAME_PREFIX
          value: {{.Values.aws.cloudformation.instanceProfileNamePrefix}}

        - key: PUBLIC_DNS_HOST_SUFFIX
          value: {{.Values.global.baseDomain}}

        - key: MSVC_TEMPLATE_FILE_PATH
          value: /infra.d/templates/managed-svc-templates.yml

      volumes:
        - mountPath: /infra.d/templates
          type: config
          refName: managed-svc-template
          items:
            - key: managed-svc-templates.yml


