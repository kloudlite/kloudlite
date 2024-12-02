{{/*{{- $obj := get . "obj" }}*/}}

{{- $name := get . "name" -}}
{{- $namespace := get . "namespace" -}}

{{- $region := get . "region" -}}

{{- $nodeSelector := get . "node-selector" | default dict -}}
{{- $tolerations := get . "tolerations" | default list -}}

{{- $ownerRefs := get . "owner-refs" | default list  -}}
{{- $labels := get . "labels" | default dict -}}
{{- $ingressClassName := get . "ingress-class-name"  -}}

{{- $wildcardCertNamespace := get . "wildcard-cert-namespace"  -}}
{{- $wildcardCertName := get . "wildcard-cert-name"  -}}
{{endl}}

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cluster-svc-account
  namespace: {{$namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{$namespace}}-ingress-rb
subjects:
  - kind: ServiceAccount
    name: cluster-svc-account
    namespace: {{$namespace}}
    apiGroup: ""
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: ""
---
{{- /*gotype: github.com/kloudlite/operator/apis/crds/v1.EdgeRouter*/ -}}
{{endl}}
apiVersion: ingress.kloudlite.io/v1
kind: Nginx
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
  labels: {{$labels | toYAML| nindent 4}}
spec:
  nameOverride: {{$name}}
  commonLabels: {{$labels  | toYAML| nindent 4}}

  rbac:
    create: false

  serviceAccount:
    create: false
    name: cluster-svc-account

  controller:
    kind: DaemonSet
    hostNetwork: true
    hostPort:
      enabled: true
      ports:
        http: 80
        https: 443
        healthz: 10254

    dnsPolicy: ClusterFirstWithHostNet

{{/*    watchIngressWithoutClass: false*/}}
    ingressClassByName: true
    ingressClass: {{$ingressClassName}}
    electionID: {{$ingressClassName}}
    ingressClassResource:
      enabled: true
      name: "{{$ingressClassName}}"
      controllerValue: "k8s.io/{{$ingressClassName}}"

    service:
      type: "ClusterIP"

    {{- if (and $wildcardCertNamespace $wildcardCertName) }}
    extraArgs:
      default-ssl-certificate: "{{$wildcardCertNamespace}}/{{$wildcardCertName}}"
    {{- end }}
    podLabels: {{$labels  | toYAML | nindent 6}}

    resources:
      requests:
        cpu: 100m
        memory: 200Mi

    nodeSelector:
      {{if $nodeSelector}}{{$nodeSelector | toYAML| nindent 6}}{{end}}
      {{include "RegionNodeSelector" (dict "region" $region) | nindent 6 }}

    tolerations:
      {{if $tolerations}}{{ $tolerations | toYAML | nindent 6}}{{end}}
      {{include "RegionToleration" (dict "region" $region) | nindent 6}}

    affinity: {{include "NodeAffinity" (dict) | nindent 6 }}

    admissionWebhooks:
      enabled: false
      failurePolicy: Ignore
