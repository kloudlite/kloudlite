{{- $namespace := get . "namespace" -}}
{{- $svcAccount := get . "svc-account" -}}
{{- $sharedConstants := get . "shared-constants" -}}

{{- $ownerRefs := get . "owner-refs" | default list  -}}
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
  name: {{.AppGqlGatewayApi}}
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
      type: tcp
      name: http
  containers:
    - name: main
      image: {{.ImageGqlGatewayApi}}
      imagePullPolicy: {{$imagePullPolicy}}
      env:
        - key: PORT
          value: '3000'
        - key: SUPERGRAPH_CONFIG
          value: /hotspot/config
      resourceCpu:
        min: 150m
        max: 300m
      resourceMemory:
        min: 200Mi
        max: 300Mi

      volumes:
        - mountPath: /hotspot
          type: config
          refName: gateway-supergraph

      livenessProbe:
        type: httpGet
        httpGet:
          path: /.well-known/apollo/server-health
          port: 3000
        initialDelay: 20
        interval: 10
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: gateway-supergraph
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
data:
  config: |+
    serviceList:
      - name: {{.AppAuthApi}}
        url: http://{{.AppAuthApi}}.{{$namespace}}.svc.cluster.local/query

      - name: {{.AppDnsApi}}
        url: http://{{.AppDnsApi}}.{{$namespace}}.svc.cluster.local/query

      - name: {{.AppCiApi}}
        url: http://{{.AppCiApi}}.{{$namespace}}.svc.cluster.local/query

      - name: {{.AppConsoleApi}}
        url: http://{{.AppConsoleApi}}.{{$namespace}}.svc.cluster.local/query

      - name: {{.AppFinanceApi}}
        url: http://{{.AppFinanceApi}}.{{$namespace}}.svc.cluster.local/query

---
apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.AppGqlGatewayApi}}
  namespace: {{$namespace}}
  labels:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  domains:
    - "gateway.{{.SubDomain}}"
  https:
    enabled: true
    forceRedirect: true
  cors:
    enabled: true
    origins:
      - https://studio.apollographql.com
    allowCredentials: true
  basicAuth:
    enabled: true
    username: {{.AppGqlGatewayApi}}
  routes:
    - app: {{.AppGqlGatewayApi}}
      path: /
      port: 80
{{end}}
