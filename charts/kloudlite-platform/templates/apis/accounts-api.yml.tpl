apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.accountsApi.name}}
  namespace: {{.Release.Namespace}}
spec:
  {{- /* region: {{.Values.region | default ""}} */}}
  serviceAccount: {{ .Values.clusterSvcAccount }}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
    - port: 3001
      targetPort: 3001
      name: grpc
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.accountsApi.image}}
      imagePullPolicy: {{.Values.apps.accountsApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "50m"
        max: "80m"
      resourceMemory:
        min: "75Mi"
        max: "100Mi"
      env:
        - key: HTTP_PORT
          value: {{.Values.apps.accountsApi.configuration.httpPort | squote}}

        - key: GRPC_PORT
          value: {{.Values.apps.accountsApi.configuration.grpcPort | squote }}

        - key: MONGO_DB_NAME
          value: {{.Values.managedResources.accountsDb}}

        - key: MONGO_URI
          type: secret
          refName: mres-{{.Values.managedResources.accountsDb}}-creds
          refKey: URI

        - key: AUTH_REDIS_HOSTS
          type: secret
          {{- /* refName: mres-{{.Values.managedResources.authRedis}} */}}
          refName: msvc-{{.Values.managedServices.redisSvc}}
          refKey: HOSTS

        - key: AUTH_REDIS_PASSWORD
          type: secret
          {{- /* refName: mres-{{.Values.managedResources.authRedis}} */}}
          refName: msvc-{{.Values.managedServices.redisSvc}}
          {{- /* refKey: PASSWORD */}}
          refKey: ROOT_PASSWORD

        - key: AUTH_REDIS_PREFIX
          {{- /* type: secret */}}
          {{- /* refName: mres-{{.Values.managedResources.authRedis}} */}}
          {{- /* refKey: PREFIX */}}
          value: auth

        - key: AUTH_REDIS_USERNAME
          value: ""
          {{- /* type: secret */}}
          {{- /* refName: mres-{{.Values.managedResources.authRedis}} */}}
          {{- /* refKey: USERNAME */}}

        - key: COOKIE_DOMAIN
          value: "{{.Values.cookieDomain}}"

        - key: IAM_GRPC_ADDR
          value: "{{.Values.apps.iamApi.name}}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}:{{.Values.apps.iamApi.configuration.grpcPort}}"

        - key: COMMS_GRPC_ADDR
          value: "{{.Values.apps.commsApi.name}}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}:{{.Values.apps.commsApi.configuration.grpcPort}}"

        - key: CONTAINER_REGISTRY_GRPC_ADDR
          value: "{{.Values.apps.containerRegistryApi.name}}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}:{{.Values.apps.containerRegistryApi.configuration.grpcPort}}"

        - key: CONSOLE_GRPC_ADDR
          value: "{{.Values.apps.consoleApi.name}}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}:{{.Values.apps.consoleApi.configuration.grpcPort}}"

        - key: AUTH_GRPC_ADDR
          value: "{{.Values.apps.authApi.name}}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}:{{.Values.apps.authApi.configuration.grpcPort}}"

