{{- if .Values.apps.containerRegistryApi.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.containerRegistryApi.name}}
  namespace: {{.Release.Namespace}}
spec:
  region: {{.Values.region | default ""}}
  serviceAccount: {{.Values.clusterSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 80
      targetPort: {{.Values.apps.containerRegistryApi.configuration.httpPort}}
      name: http
      type: tcp

    - port: 4001
      targetPort: {{.Values.apps.containerRegistryApi.configuration.grpcPort}}
      name: grpc
      type: tcp

    - port: {{.Values.apps.containerRegistryApi.configuration.grpcPort}}
      targetPort: {{.Values.apps.containerRegistryApi.configuration.grpcPort}}
      name: grpc
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.containerRegistryApi.image}}
      imagePullPolicy: {{.Values.apps.containerRegistryApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "30m"
        max: "50m"
      resourceMemory:
        min: "50Mi"
        max: "80Mi"
      env:
        - key: PORT
          value: {{.Values.apps.containerRegistryApi.configuration.httpPort | squote}}

        - key: GRPC_PORT
          value: {{.Values.apps.containerRegistryApi.configuration.grpcPort | squote}}

        - key: REGISTRY_EVENT_LISTNER_PORT
          value: {{.Values.apps.containerRegistryApi.configuration.eventListenerPort | squote}}

        - key: REGISTRY_AUTHORIZER_PORT
          value: {{.Values.apps.containerRegistryApi.configuration.authorizerPort | squote}}

        - key: REGISTRY_SECRET_KEY
          value: {{.Values.apps.containerRegistryApi.configuration.registrySecret | squote}}

        - key: COOKIE_DOMAIN
          value: {{.Values.cookieDomain}}

        - key: ACCOUNT_COOKIE_NAME
          value: {{.Values.accountCookieName}}

        - key: AUTH_REDIS_HOSTS
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: HOSTS

        - key: AUTH_REDIS_PASSWORD
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: PASSWORD

        - key: AUTH_REDIS_PREFIX
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: PREFIX

        - key: AUTH_REDIS_USERNAME
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: USERNAME

        - key: REGISTRY_URL
          value: http://{{(index .Values.helmCharts "container-registry").name }}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}

        - key: REGISTRY_REDIS_USERNAME
          type: secret
          refName: mres-{{.Values.managedResources.containerRegistryRedis}}
          refKey: USERNAME

        - key: REGISTRY_REDIS_PREFIX
          type: secret
          refName: mres-{{.Values.managedResources.containerRegistryRedis}}
          refKey: PREFIX

        - key: REGISTRY_REDIS_HOSTS
          type: secret
          refName: mres-{{.Values.managedResources.containerRegistryRedis}}
          refKey: HOSTS

        - key: REGISTRY_REDIS_PASSWORD
          type: secret
          refName: mres-{{.Values.managedResources.containerRegistryRedis}}
          refKey: PASSWORD

        - key: DB_URI
          type: secret
          refName: mres-{{.Values.managedResources.containerRegistryDb}}
          refKey: URI

        - key: DB_NAME
          value: {{.Values.managedResources.containerRegistryDb}}

        - key: IAM_GRPC_ADDR
          value: {{.Values.apps.iamApi.name}}.{{.Release.Namespace}}.svc.{{.Values.clusterInternalDNS}}:{{.Values.apps.iamApi.configuration.grpcPort}}
---
{{- end }}
