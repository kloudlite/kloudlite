{{- $namespace := get . "namespace" | default "kl-core" -}}
{{- $accountRef := get . "account-ref" | default "kl-core" -}}
{{- $region := get . "region" | default "master" -}}
{{- $imageAuthApi := get . "image-auth-api" -}}
{{- $imagePullPolicy := get . "image-pull-policy" | default "Always" -}}

{{- $sharedConstants := get . "shared-constants" -}}

{{ with $sharedConstants}}
{{/*gotype: operators.kloudlite.io/apis/cluster-setup/v1.SharedConstants*/}}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.AppAuthApi}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
spec:
  region: {{$region}}
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
      image: {{$imageAuthApi}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"
      env:
        - key: MONGO_DB_NAME
          value: {{.AuthDbName}}

        - key: REDIS_HOSTS
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: mres-{{.AuthRedisName}}
          refKey: USERNAME

        - key: MONGO_URI
          type: secret
          refName: mres-{{.AuthDbName}}
          refKey: URI

        - key: COMMS_SERVICE
          value: "{{.AppCommsApi}}.{{$namespace}}.svc.cluster.local:3001"

        - key: COMMS_HOST
          value: {{.AppCommsApi}}

        - key: COMMS_PORT
          value: "3001"

        - key: PORT
          value: "3000"

        - key: GRPC_PORT
          value: "3001"

        - key: ORIGINS
          value: "http://localhost:4001,https://studio.apollographql.com"

        - key: COOKIE_DOMAIN
          value: "{{.CookieDomain}}"

        - key: GITHUB_APP_PK_FILE
          value: /github/github-app-pk.pem

      envFrom:
        - type: secret
          refName: oauth-secrets

      volumes:
        - mountPath: /github
          type: secret
          refName: oauth-github-app-pk
          items:
            - key: github-app-pk.pem
              fileName: github-app-pk.pem
{{ end}}
