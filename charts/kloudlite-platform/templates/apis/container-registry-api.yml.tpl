{{- if .Values.apps.containerRegistryApi.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: App
metaata:
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

        - key: COOKIE_DOMAIN
          value: {{.Values.cookieDomain}}

        - key: ACCOUNT_COOKIE_NAME
          value: {{.Values.accountCookieName}}

        - key: HARBOR_REGISTRY_HOST
          type: secret
          refName: {{.Values.secretNames.harborAdminSecret}}
          refKey: IMAGE_REGISTRY_HOST

        - key: HARBOR_ADMIN_USERNAME
          type: secret
          refName: {{.Values.secretNames.harborAdminSecret}}
          refKey: ADMIN_USERNAME

        - key: HARBOR_ADMIN_PASSWORD
          type: secret
          refName: {{.Values.secretNames.harborAdminSecret}}
          refKey: ADMIN_PASSWORD

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

        - key: DB_URI
          type: secret
          refName: mres-{{.Values.managedResources.containerRegistryDb}}
          refKey: URI

        - key: DB_NAME
          value: {{.Values.managedResources.containerRegistryDb}}
---
{{- end }}
