{{- $appName := include "apps.infraApi.name" . }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{$appName}}
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4}}
  annotations:
    kloudlite.io/checksum.edge-gateways: {{ include (print $.Template.BasePath "/03-apps/apis/secrets/edge-gateways.yml.tpl") . | sha256sum }}
    {{ include "common.pod-annotations" . | nindent 4}}
spec:
  serviceAccount: {{.Values.serviceAccounts.clusterAdmin.name}}

  nodeSelector: {{.Values.scheduling.stateless.nodeSelector | toYaml | nindent 4}}
  tolerations: {{.Values.scheduling.stateless.tolerations | toYaml | nindent 4}}
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  replicas: {{.Values.apps.infraApi.replicas}}

  services:
    - port: {{ include "apps.infraApi.httpPort" . }}
    - port: {{ include "apps.infraApi.grpcPort" . }}

  hpa:
    enabled: {{ .Values.apps.infraApi.hpa.enabled }}
    minReplicas: {{.Values.apps.infraApi.hpa.minReplicas}}
    maxReplicas: {{.Values.apps.infraApi.hpa.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: '{{.Values.apps.infraApi.image.repository}}:{{.Values.apps.infraApi.image.tag | default (include "image-tag" .) }}'
      imagePullPolicy: {{ include "image-pull-policy" .}}
      
      resourceCpu:
        min: "50m"
        max: "200m"
      resourceMemory:
        min: "50Mi"
        max: "200Mi"

      env:
        - key: HTTP_PORT
          value: {{ include "apps.infraApi.httpPort" . | squote }}

        - key: GRPC_PORT
          value: {{ include "apps.infraApi.grpcPort" . | squote }}

        - key: ACCOUNTS_GRPC_ADDR
          value: 'accounts-api:{{ include "apps.accountsApi.grpcPort" . }}'

        - key: MONGO_DB_URI
          type: secret
          refName: mres-infra-db-creds
          refKey: .CLUSTER_LOCAL_URI

        - key: MONGO_DB_NAME
          type: secret
          refName: mres-infra-db-creds
          refKey: DB_NAME

        - key: KLOUDLITE_DNS_SUFFIX
          value: {{ include "kloudlite.dns-suffix" . }}

        - key: COOKIE_DOMAIN
          value: "{{- include "kloudlite.cookie-domain" . }}"

        - key: NATS_URL
          value: {{ include "nats.url" . }}

        - key: ACCOUNT_COOKIE_NAME
          value: {{ include "kloudlite.account-cookie-name" . }}

        - key: PROVIDER_SECRET_NAMESPACE
          value: {{.Release.Namespace}}

        - key: IAM_GRPC_ADDR
          value: '{{ include "apps.iamApi.name" . }}:{{ include "apps.iamApi.grpcPort" . }}'

        - key: CONSOLE_GRPC_ADDR
          value: '{{ include "apps.consoleApi.name" . }}:{{ include "apps.consoleApi.grpcPort" . }}'

        - key: NATS_RECEIVE_FROM_AGENT_STREAM
          value: {{.Values.nats.streams.receiveFromAgent.name}}

        - key: SESSION_KV_BUCKET
          value: {{.Values.nats.buckets.sessionKVBucket.name}}

        - key: MESSAGE_OFFICE_INTERNAL_GRPC_ADDR
          value: "{{ include "apps.messageOffice.name" . }}:{{ include "apps.messageOffice.privateGrpcPort" . }}"

        - key: MESSAGE_OFFICE_EXTERNAL_GRPC_ADDR
          value: 'message-office.{{.Values.baseDomain }}:443'

        - key: KLOUDLITE_RELEASE
          value: {{.Values.apps.infraApi.kloudliteRelease | default .Values.kloudliteRelease}}

        - key: MSVC_TEMPLATE_FILE_PATH
          value: /infra.d/templates/managed-svc-templates.yml

        - key: GLOBAL_VPN_KUBE_REVERSE_PROXY_IMAGE
          value: '{{.Values.apps.infraApi.imageGatewayKubeProxy.repository}}:{{.Values.apps.infraApi.imageGatewayKubeProxy.tag | default (include "image-tag" .)}}'

        - key: GLOBAL_VPN_KUBE_REVERSE_PROXY_AUTHZ_TOKEN
          type: secret
          refName: {{ include "apps.gatewayKubeReverseProxy.secret.name" . }}
          refKey: {{ include "apps.gatewayKubeReverseProxy.secret.key" . }}

        - key: KLOUDLITE_GLOBAL_VPN_DEVICE_HOST
          value: wg-gateways.{{.Values.baseDomain}}

        - key: AVAILABLE_KLOUDLITE_REGIONS_CONFIG
          value: "/kloudlite/gateways.yml"

        - key: KLOUDLITE_EDGE_GATEWAY_SERVICE_TYPE
          value: {{.Values.apps.infraApi.edgeGatewayServiceType }}

      volumes:
        - mountPath: /infra.d/templates
          type: config
          refName: managed-svc-template
          items:
            - key: managed-svc-templates.yml

        - mountPath: /kloudlite
          type: secret
          refName: {{ include "edge-gateways.secret.name" . }}
          items:
            - key: {{ include "edge-gateways.secret.key" . }}
              fileName: gateways.yml

      livenessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.infraApi.httpPort" . }}
        initialDelay: 10
        interval: 10

      readinessProbe:
        type: httpGet
        httpGet:
          path: /_healthy
          port: {{ include "apps.infraApi.httpPort" . }}
        initialDelay: 10
        interval: 10
