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
  name: {{.AppCommsApi}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerREference: {{$ownerRefs  | toYAML | nindent 4}}
spec:
  region: {{$region}}
  nodeSelector: {{$nodeSelector | toYAML | nindent 4}}
  tolerations: {{$tolerations | toYAML | nindent 4}}
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
      image: {{.ImageCommsApi}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"

      env:
        - key: GRPC_PORT
          value: "3001"

        - key: SUPPORT_EMAIL
          value: support@kloudlite.io

        - key: SENDGRID_API_KEY
          value: '***REMOVED***'

        - key: EMAIL_LINKS_BASE_URL
          value: https://auth.{{.SubDomain}}/

      # envFrom:
      #   - type: secret
      #     refName: comms-env

{{end}}
