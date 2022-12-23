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
  name: {{.AppJsEvalApi}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML| nindent 4}}
spec:
  region: {{$region}}
  nodeSelector: {{$nodeSelector | toYAML | nindent 4}}
  tolerations: {{$tolerations | toYAML | nindent 4}}
  services:
    - port: 3001
      targetPort: 3001
      name: grpc
      type: tcp
  containers:
    - name: main
      image: {{.ImageJsEvalApi}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "25m"
        max: "50m"
      resourceMemory:
        min: "50Mi"
        max: "100Mi"
      env:
        - key: PORT
          value: "3001"
{{end}}
