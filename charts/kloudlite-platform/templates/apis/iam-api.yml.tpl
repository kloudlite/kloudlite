apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.iamApi.name}}
  namespace: {{.Release.Namespace}}
spec:
  region: {{.Values.region | default ""}}
  serviceAccount: {{.Values.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: {{.Values.apps.iamApi.configuration.grpcPort}}
      targetPort: {{.Values.apps.iamApi.configuration.grpcPort}}
      name: grpc
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.iamApi.image}}
      imagePullPolicy: {{.Values.apps.iamApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      
      resourceCpu:
        min: "30m"
        max: "50m"
      resourceMemory:
        min: "50Mi"
        max: "100Mi"
      env:
        - key: MONGO_DB_NAME
          value: {{.Values.managedResources.iamDb}}

        - key: REDIS_HOSTS
          type: secret
          {{- /* refName: mres-{{.Values.managedResources.iamRedis}} */}}
          {{- /* refKey: HOSTS */}}
          refName: msvc-{{.Values.managedServices.redisSvc}}
          refKey: HOSTS


        - key: REDIS_PASSWORD
          type: secret
          {{- /* refName: mres-{{.Values.managedResources.iamRedis}} */}}
          {{- /* refKey: PASSWORD */}}
          refName: msvc-{{.Values.managedServices.redisSvc}}
          refKey: ROOT_PASSWORD

        - key: REDIS_PREFIX
          value: "iam"
          {{- /* type: secret */}}
          {{- /* refName: mres-{{.Values.managedResources.iamRedis}} */}}
          {{- /* refKey: PREFIX */}}

        - key: REDIS_USERNAME
          value: ""
          {{- /* type: secret */}}
          {{- /* refName: mres-{{.Values.managedResources.iamRedis}} */}}
          {{- /* refKey: USERNAME */}}

        - key: MONGO_DB_URI
          type: secret
          refName: mres-{{.Values.managedResources.iamDb}}-creds
          refKey: URI

        - key: COOKIE_DOMAIN
          value: "{{.Values.cookieDomain}}"

        - key: GRPC_PORT
          value: {{.Values.apps.iamApi.configuration.grpcPort | squote}}

        - key: CONSOLE_SERVICE
          value: "{{.Values.apps.consoleApi.name}}:{{.Values.apps.consoleApi.configuration.grpcPort}}"
