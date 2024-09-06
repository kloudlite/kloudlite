{{- $obj := get . "object"}}
{{- $volumes := get . "volumes"}}
{{- $vMounts := get . "volume-mounts"}}
{{- $ownerRefs := get . "owner-refs" | default list  }}

{{- $podLabels := get . "pod-labels" | default dict }}
{{- $podAnnotations := get . "pod-annotations" | default dict }}

{{- /* {{- $clusterDnsSuffix := get . "cluster-dns-suffix" | default "cluster.local"}} */}}

{{- with $obj }}

{{- $isIntercepted := (and .Spec.Intercept .Spec.Intercept.Enabled) }}
{{- $isHpaEnabled := (and .Spec.Hpa .Spec.Hpa.Enabled) }}

{{- /* gotype: github.com/kloudlite/operator/apis/crds/v1.App */ -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
  labels: {{.Labels | toYAML | nindent 4}}
spec:
  {{- if (or .Spec.Freeze $isIntercepted) }}
  replicas: 0
  {{- else if $isHpaEnabled }} 
  {{- /* when hpa-enabled, we don't want to set replicas */}}
  {{- else }}
  replicas: {{.Spec.Replicas}}
  {{- end }}
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
        {{ $podLabels | toYAML | nindent 8 }}
      annotations: {{$podAnnotations | toYAML | nindent 8 }}
    spec:
      serviceAccountName: {{.Spec.ServiceAccount}}
      nodeSelector: {{if .Spec.NodeSelector}}{{ .Spec.NodeSelector | toYAML | nindent 8 }}{{end}}
        {{- if .Spec.Region}}
        kloudlite.io/region: {{.Spec.Region | squote}}
        {{- end}}

      tolerations: {{if .Spec.Tolerations}}{{.Spec.Tolerations | toYAML | nindent 8}}{{end}}
        {{- if .Spec.Region}}
        - effect: NoExecute
          key: kloudlite.io/region
          operator: Equal
          value: {{.Spec.Region | squote}}
        {{- end}}

      {{- if .Spec.TopologySpreadConstraints }}
      topologySpreadConstraints: {{.Spec.TopologySpreadConstraints | toYAML | nindent 8}}
      {{- end }}

      dnsPolicy: ClusterFirst

      {{- /* affinity: */ -}}
      {{- /*   nodeAffinity: */ -}}
      {{- /*     preferredDuringSchedulingIgnoredDuringExecution: */ -}}
      {{- /*       {{- $nWeight := 30 -}} */ -}}
      {{- /*       {{- range $weight := Iterate $nWeight }} */ -}}
      {{- /*       - weight: {{ sub $nWeight $weight }} */ -}}
      {{- /*         preference: */ -}}
      {{- /*           matchExpressions: */ -}}
      {{- /*             - key: kloudlite.io/node-index */ -}}
      {{- /*               operator: In */ -}}
      {{- /*               values: */ -}}
      {{- /*                 - {{$weight | squote}} */ -}}
      {{- /*       {{- end }} */ -}}

      {{- if .Spec.Containers }}
      {{- $myDict := dict "containers" .Spec.Containers "volumeMounts" $vMounts }}
      containers: {{- include "TemplateContainer" $myDict | nindent 8 }}
      {{- if $volumes }}
      volumes: {{- $volumes| toYAML | nindent 8 }}
      {{- end }}
      {{- end }}
---
{{- if .Spec.Services }}
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
spec:
  selector: {{ $podLabels | toYAML | nindent 4}}
  ports:
    {{- range $svc := .Spec.Services }}
    {{- with $svc }}
    - protocol: {{.Protocol | default "TCP"}}
      port: {{.Port}}
      name: {{printf "p-%d" .Port | squote}}
      targetPort: {{.Port}}
    {{- end }}
    {{- end }}
{{- end }}
{{- end }}
