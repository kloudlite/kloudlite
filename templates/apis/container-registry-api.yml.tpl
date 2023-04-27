apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.containerRegistryApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region | default ""}}
  serviceAccount: {{.Values.clusterSvcAccount}}

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
          value: "3000"

        - key: GRPC_PORT
          value: "3001"

        - key: COOKIE_DOMAIN
          value: {{.Values.cookieDomain}}

        - key: ACCOUNT_COOKIE_NAME
          value: {{.Values.accountCookieName}}

        - key: HARBOR_REGISTRY_HOST
          type: secret
          refName: {{.Values.secrets.names.harborAdminSecret}}
          refKey: IMAGE_REGISTRY_HOST

        - key: HARBOR_ADMIN_USERNAME
          type: secret
          refName: {{.Values.secrets.names.harborAdminSecret}}
          refKey: ADMIN_USERNAME

        - key: HARBOR_ADMIN_PASSWORD
          type: secret
          refName: {{.Values.secrets.names.harborAdminSecret}}
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
