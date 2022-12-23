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
  name: {{.AppAuthWeb}}
  namespace: {{$namespace}}
  labels:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML| nindent 4}}
spec:
  region: {{$region}}
  nodeSelector: {{$nodeSelector | toYAML| nindent 4}}
  tolerations: {{$tolerations | toYAML| nindent 4}}
  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
  containers:
    - name: main
      image: {{.ImageAuthWeb}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "200m"
        max: "300m"
      resourceMemory:
        min: "200Mi"
        max: "300Mi"
      env:
        - key: BASE_URL
          value: "{{.SubDomain}}"
        - key: ENV
          value: "development"
        - key: PORT
          value: "3000"
---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.AppAuthWeb}}
  namespace: {{$namespace}}
  labels:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML| nindent 4}}
spec:
  domains:
    - "auth.{{.SubDomain}}"
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: {{.AppAuthWeb}}
      path: /
      port: 80
{{end}}
