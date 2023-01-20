{{- $namespace := get . "namespace" -}}
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
  name: {{.AppConsoleWeb}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs| toYAML | nindent 4}}
spec:
  region: {{$region}}
  nodeSelector: {{$nodeSelector| toYAML | nindent 4}}
  tolerations: {{$tolerations | toYAML | nindent 4}}
  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
  containers:
    - name: main
      image: {{.ImageConsoleWeb}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "200m"
        max: "300m"
      resourceMemory:
        min: "200Mi"
        max: "300Mi"
      livenessProbe: &probe
        type: httpGet
        initialDelay: 5
        failureThreshold: 3
        httpGet:
          path: /console/assets/healthy.txt
          port: 3000
        interval: 10
      readinessProbe: *probe
      env:
        - key: BASE_URL
          value: {{.SubDomain}}
        - key: ENV
          value: "development"
        - key: PORT
          value: "3000"
        - key: GITHUB_APP
          value: "kloudlite-dev"
---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.AppConsoleWeb}}
  namespace: {{$namespace}}
  labels:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs |toYAML | nindent 4}}
spec:
  domains:
    - "console.{{.SubDomain}}"
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: {{.AppConsoleWeb}}
      path: /
      port: 80
{{end}}
