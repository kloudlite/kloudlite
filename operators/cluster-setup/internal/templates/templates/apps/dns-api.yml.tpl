{{- $namespace := get . "namespace" -}}
{{- $svcAccount := get . "svc-account" -}}
{{- $sharedConstants := get . "shared-constants" -}}

{{- $ownerRefs := get . "owner-refs" | default list -}}
{{- $accountRef := get . "account-ref" | default "kl-core" -}}
{{- $region := get . "region" | default "master" -}}
{{- $imagePullPolicy := get . "image-pull-policy" | default "Always" -}}

{{- $nodeSelector := get . "node-selector" | default dict -}}
{{- $tolerations := get . "tolerations" | default list -}}

{{- $dnsNames := get . "dns-names" -}}
{{- $cnameBaseDomain := get . "cname-base-domain" -}}

{{ with $sharedConstants}}
{{/*gotype: operators.kloudlite.io/apis/cluster-setup/v1.SharedConstants*/}}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.AppDnsApi}}
  namespace: {{$namespace}}
  annotations:
    kloudlite.io/account-ref: {{$accountRef}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  region: {{$region}}
  nodeSelector: {{$nodeSelector | toYAML | nindent 4}}
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
      image: {{.ImageDnsApi}}
      imagePullPolicy: {{$imagePullPolicy}}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "100Mi"
        max: "200Mi"
      env:
        - key: DNS_DOMAIN_NAMES
          value: "{{$dnsNames}}"

        - key: EDGE_CNAME_BASE_DOMAIN
          value: "{{$cnameBaseDomain}}"

        - key: REDIS_HOSTS
          type: secret
          refName: mres-{{.DnsRedisName}}
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: mres-{{.DnsRedisName}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: mres-{{.DnsRedisName}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: mres-{{.DnsRedisName}}
          refKey: USERNAME

        - key: MONGO_URI
          type: secret
          refName: mres-{{.DnsDbName}}
          refKey: URI

        - key: MONGO_DB_NAME
          value: {{.DnsDbName}}

        - key: CONSOLE_SERVICE
          value: "{{.AppConsoleApi}}.{{$namespace}}.svc.cluster.local:3001"

        - key: FINANCE_SERVICE
          value: "{{.AppFinanceApi}}.{{$namespace}}.svc.cluster.local:3001"

        - key: PORT
          value: '3000'

        - key: GRPC_PORT
          value: '3001'

        - key: DNS_PORT
          value: '5353'

---

apiVersion: v1
kind: Service
metadata:
  name: {{.AppDnsApi}}-exposed
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs |toYAML | nindent 4}}
spec:
  ports:
    - name: server-udp
      nodePort: 30053
      port: 5353
      protocol: UDP
      targetPort: 5353
  selector:
    app: {{.AppDnsApi}}
  type: NodePort

---

apiVersion: v1
kind: Secret
metadata:
  name: {{.AppDnsApi}}-basic-auth
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs |toYAML | nindent 4}}
data:
  auth: ***REMOVED***

---

apiVersion: crds.kloudlite.io/v1
kind: Router
metadata:
  name: {{.AppDnsApi}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs |toYAML | nindent 4}}
spec:
  domains:
    - "{{.AppDnsApi}}.{{.SubDomain}}"
  https:
    enabled: true
    forceRedirect: true
  basicAuth:
    enabled: true
    secretName: {{.AppDnsApi}}-basic-auth
  routes:
    - app: {{.AppDnsApi}}
      path: /
      port: 80
{{ end }}
