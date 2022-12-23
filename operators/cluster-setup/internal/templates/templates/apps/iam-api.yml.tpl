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
{{/*gotype: operators.kloudlite.io/apis/cluster-setup/v1.SharedConstants*/}}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.AppIAMApi}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  region: {{$region}}
  nodeSelector: {{$nodeSelector | toYAML |nindent 4}}
  tolerations: {{$tolerations | toYAML |nindent 4}}
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
      image: {{.ImageIAMApi}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"
      env:

        - key: MONGO_DB_NAME
          value: {{.IamDbName}}

        - key: REDIS_HOSTS
          type: secret
          refName: mres-{{.IamRedisName}}
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: mres-{{.IamRedisName}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: mres-{{.IamRedisName}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: mres-{{.IamRedisName}}
          refKey: USERNAME

        - key: MONGO_DB_URI
          type: secret
          refName: mres-{{.IamDbName}}
          refKey: URI

        - key: COOKIE_DOMAIN
          value: ".{{.CookieDomain}}"

        - key: PORT
          value: "3000"

        - key: GRPC_PORT
          value: "3001"

        - key: CONSOLE_SERVICE
          value: "{{.AppConsoleApi}}:3001"
{{end}}
