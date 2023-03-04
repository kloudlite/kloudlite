apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.iamApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region}}

  {{- if .Values.nodeSelector}}
  nodeSelector: {{.Values.nodeSelector | toYaml | nindent 4}}
  {{- end }}

  {{- if .Values.tolerations }}
  tolerations: {{.Values.tolerations | toYaml | nindent 4}}
  {{- end }}

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
          refName: mres-{{.Values.managedResources.iamRedis}}
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: mres-{{.Values.managedResources.iamRedis}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: mres-{{.Values.managedResources.iamRedis}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: mres-{{.Values.managedResources.iamRedis}}
          refKey: USERNAME

        - key: MONGO_DB_URI
          type: secret
          refName: mres-{{.Values.managedResources.iamDb}}
          refKey: URI

        - key: COOKIE_DOMAIN
          value: "{{.Values.cookieDomain}}"

        - key: PORT
          value: "3000"

        - key: GRPC_PORT
          value: "3001"

        - key: CONSOLE_SERVICE
          value: "{{.Values.apps.consoleApi.name}}:3001"
