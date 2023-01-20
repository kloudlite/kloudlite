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
  name: {{.AppSocketWeb}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  region: {{$region}}
  services:
    - port: 80
      targetPort: 3000
      name: socket
      type: tcp
    - port: 3001
      targetPort: 3001
      name: http
      type: tcp

  containers:
    - name: main
      image: {{.ImageSocketWeb}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "200Mi"
        max: "200Mi"
      env:
        - key: BASE_URL
          value: {{.SubDomain}}
        - key: ENV
          value: "development"
        - key: REDIS_URI
          type: secret
          refName: mres-{{.SocketRedisName}}
          refKey: URI

---

apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.AppSocketWeb}}
  namespace: {{$namespace}}
  labels:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  domains:
    - "socket.{{.SubDomain}}"
  https:
    enabled: true
    forceRedirect: true
  routes:
    - app: {{.AppSocketWeb}}
      path: /
      port: 80
    - app: {{.AppSocketWeb}}
      path: /publish
      port: 3001
{{end}}
