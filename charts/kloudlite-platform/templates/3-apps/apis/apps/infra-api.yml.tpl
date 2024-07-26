{{- $appName := "infra-api" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: infra-api
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.clusterSvcAccount}}

  nodeSelector: {{include "stateless-node-selector" . | nindent 4 }}
  tolerations: {{include "stateless-tolerations" . | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}
    {{ include "tsc-nodepool" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.consoleApi.configuration.replicas}}

  services:
    - port: 3000
    - port: 3001
  containers:
    - name: main
      image: {{.Values.apps.infraApi.image.repository}}:{{.Values.apps.infraApi.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
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
          refKey: .CLUSTER_LOCAL_URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-infra-db-creds
          refKey: DB_NAME

        - key: HTTP_PORT
          value: "3000"

        - key: GRPC_PORT
          value: "3001"

        - key: KLOUDLITE_DNS_SUFFIX
          value: "{{.Values.global.kloudliteDNSSuffix}}"

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

        - key: CONSOLE_GRPC_ADDR
          value: "console-api:3001"

        - key: NATS_RECEIVE_FROM_AGENT_STREAM
          value: {{.Values.envVars.nats.streams.receiveFromAgent.name}}

        - key: SESSION_KV_BUCKET
          value: {{.Values.envVars.nats.buckets.sessionKVBucket.name}}

        - key: MESSAGE_OFFICE_INTERNAL_GRPC_ADDR
          value: "message-office:3002"

        - key: MESSAGE_OFFICE_EXTERNAL_GRPC_ADDR
          value: message-office.{{include "router-domain" . }}:443

        - key: KLOUDLITE_RELEASE
          value: {{.Values.apps.infraApi.configuration.kloudliteRelease | default .Values.global.kloudlite_release}}

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

        - key: GLOBAL_VPN_KUBE_REVERSE_PROXY_IMAGE
          value: ghcr.io/kloudlite/api/cmds/global-vpn-kube-proxy:v1.0.7-nightly

        - key: GLOBAL_VPN_KUBE_REVERSE_PROXY_AUTHZ_TOKEN
          value: {{.Values.apps.infraApi.configuration.globalVpnKubeReverseProxyAuthzToken}}

        - key: KLOUDLITE_GLOBAL_VPN_DEVICE_HOST
          value: wg-gateways.{{.Values.global.baseDomain}}

        - key: AVAILABLE_KLOUDLITE_REGIONS_CONFIG
          value: "/kloudlite/gateways.yml"

      volumes:
        - mountPath: /infra.d/templates
          type: config
          refName: managed-svc-template
          items:
            - key: managed-svc-templates.yml

        - mountPath: /kloudlite
          type: secret
          refName: {{.Values.edgeGateways.secretKeyRef.name}}
          items:
            - key: {{.Values.edgeGateways.secretKeyRef.key}}
              fileName: gateways.yml

      livenessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{.Values.apps.infraApi.configuration.httpPort}}
        initialDelay: 5
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{.Values.apps.infraApi.configuration.httpPort}}
        initialDelay: 5
        interval: 10
