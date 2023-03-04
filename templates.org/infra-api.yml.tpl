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
  name: {{.AppInfraApi}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  region: {{$region}}
  serviceAccount: {{$svcAccount}}
  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp

  containers:
    - name: main
      image: {{.ImageInfraApi}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "50m"
        max: "50m"
      resourceMemory:
        min: "50Mi"
        max: "50Mi"

      env:
        - key: FINANCE_GRPC_ADDR
          value: http://{{.AppFinanceApi}}:3001

        - key: INFRA_DB_NAME
          value: {{.InfraDbName}}

        - key: INFRA_DB_URI
          type: secret
          refName: "mres-{{.InfraDbName}}"
          refKey: URI

        - key: HTTP_PORT
          value: "3000"

        - key: COOKIE_DOMAIN
          value: "{{.CookieDomain}}"

        - key: AUTH_REDIS_HOSTS
          type: secret
          refName: "mres-{{.AuthRedisName}}"
          refKey: HOSTS

        - key: AUTH_REDIS_PASSWORD
          type: secret
          refName: "mres-{{.AuthRedisName}}"
          refKey: PASSWORD

        - key: AUTH_REDIS_PREFIX
          type: secret
          refName: "mres-{{.AuthRedisName}}"
          refKey: PREFIX

        - key: AUTH_REDIS_USER_NAME
          type: secret
          refName: "mres-{{.AuthRedisName}}"
          refKey: USERNAME

{{end}}
