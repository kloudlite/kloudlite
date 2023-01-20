{{- $obj := get . "obj" }}
{{- $ownerRefs := get . "owner-refs" }}
{{- $controllerName := get . "controller-name" }}
{{- $labels := get . "labels" }}
{{- $ingressClassName := get . "ingress-class-name"  -}}
{{- $clusterServiceAccountName := get . "cluster-service-account-name" -}}

{{- $wildcardCertNamespace := get . "wildcard-cert-namespace"  -}}
{{- $wildcardCertName := get . "wildcard-cert-name"  -}}

{{- with $obj }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cluster-svc-account
  namespace: {{.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: {{.Namespace}}-ingress-rb
subjects:
  - kind: ServiceAccount
    name: cluster-svc-account
    namespace: {{.Namespace}}
    apiGroup: ""
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: ""
---
{{- /*gotype: github.com/kloudlite/operator/apis/crds/v1.EdgeRouter*/ -}}
{{""}}
apiVersion: ingress.kloudlite.io/v1
kind: Nginx
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
  labels: {{$labels | toYAML| nindent 4}}
spec:
  nameOverride: {{.Name}}
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

    extraArgs:
      default-ssl-certificate: "{{$wildcardCertNamespace}}/{{$wildcardCertName}}"
    podLabels: {{$labels  | toYAML | nindent 6}}

    resources:
      requests:
        cpu: 100m
        memory: 200Mi

    nodeSelector:
      {{if .Spec.NodeSelector}}{{.Spec.NodeSelector | toYAML| nindent 6}}{{end}}
      {{include "RegionNodeSelector" (dict "region" .Spec.EdgeName) | nindent 6 }}

    tolerations:
      {{if .Spec.Tolerations}}{{.Spec.Tolerations | toYAML | nindent 6}}{{end}}
      {{include "RegionToleration" (dict "region" .Spec.EdgeName) | nindent 6}}

    affinity: {{include "NodeAffinity" (dict) | nindent 6 }}

    admissionWebhooks:
      enabled: false
      failurePolicy: Ignore
{{- end }}
