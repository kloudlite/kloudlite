{{- $namespace := get . "namespace" -}}
{{- $svcAccount := get . "svc-account" -}}
{{- $sharedConstants := get . "shared-constants" -}}

{{- $ownerRefs := get . "owner-refs" | default list -}}
{{- $accountRef := get . "account-ref" | default "kl-core" -}}
{{- $region := get . "region" | default "master" -}}
{{- $imagePullPolicy := get . "image-pull-policy" | default "Always" -}}

{{- $nodeSelector := get . "node-selector" | default dict -}}
{{- $tolerations := get . "tolerations" | default list -}}

{{ with $sharedConstants}}
{{/*gotype: github.com/kloudlite/operator/apis/cluster-setup/v1.SharedConstants*/}}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.authApi.name}}
  namespace: {{.Release.Namespace}}
  annotations:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  region: {{.Values.region}}
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
      image: {{.Values.authApi.image}}
      imagePullPolicy: {{.Values.authApi.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"
      env:
        - key: MONGO_DB_NAME
          value: {{.Values.managedResources.authDb}}

        - key: REDIS_HOSTS
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: mres-{{.Values.managedResources.authRedis}}
          refKey: USERNAME

        - key: MONGO_URI
          type: secret
          refName: mres-{{.Values.managedResources.authDb}}
          refKey: URI

        - key: COMMS_SERVICE
          value: "{{.Values.apps.commsApi.name}}.{{.Release.Namespace}}.svc.cluster.local:3001"

        - key: COMMS_HOST
          value: {{.Values.apps.commsApi.name}}

        - key: COMMS_PORT
          value: "3001"

        - key: PORT
          value: "3000"

        - key: GRPC_PORT
          value: "3001"

        - key: ORIGINS
          {{/* value: "https://{{.AuthWebDomain}},http://localhost:4001,https://studio.apollographql.com" */}}
          value: "https://kloudlite.io,http://localhost:4001,https://studio.apollographql.com"

        - key: COOKIE_DOMAIN
          value: "{{.Values.cookieDomain}}"

        - key: GITHUB_APP_PK_FILE
          value: /github/github-app-pk.pem

      envFrom:
        - type: secret
          refName: {{.Values.oAuthSecretName}}

      volumes:
        - mountPath: /github
          type: secret
          refName: oauth-secrets
          items:
            - key: github-app-pk.pem
              fileName: github-app-pk.pem
{{ end}}
