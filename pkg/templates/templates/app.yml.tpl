{{- $obj := get . "object"}}
{{- $volumes := get . "volumes"}}
{{- $vMounts := get . "volume-mounts"}}
{{- $ownerRefs := get . "owner-refs" | default list  }}
{{ $accountName := get . "account-name"}} 

{{- $clusterDnsSuffix := get . "cluster-dns-suffix" | default "svc.cluster.local"}}

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
  {{- if not (and .Spec.Hpa .Spec.Hpa.Enabled) }}
  replicas: {{if (or .Spec.Freeze $isIntercepted )}}0{{ else }}{{.Spec.Replicas}}{{end}}
  {{- end}}
  selector:
    matchLabels:
      app: {{.Name}}
  template:
    metadata:
      labels:
        app: {{.Name}}
        kloudlite.io/region: {{.Spec.Region}}
    spec:
      serviceAccount: {{.Spec.ServiceAccount}}
      nodeSelector: {{if .Spec.NodeSelector}}{{ .Spec.NodeSelector | toYAML | nindent 8 }}{{end}}
        kloudlite.io/region: {{.Spec.Region | squote}}

      tolerations: {{if .Spec.Tolerations}}{{.Spec.Tolerations | toYAML | nindent 8}}{{end}}
        - effect: NoExecute
          key: kloudlite.io/region
          operator: Equal
          value: {{.Spec.Region | squote}}

      dnsPolicy: ClusterFirst

      affinity:
        nodeAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            {{- $nWeight := 30 -}}
            {{- range $weight := Iterate $nWeight }}
            - weight: {{ sub $nWeight $weight }}
              preference:
                matchExpressions:
                  - key: kloudlite.io/node-index
                    operator: In
                    values:
                      - {{$weight | squote}}
            {{- end }}

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
  namespace: {{.Namespace}}
  name: {{.Name}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
spec:
  type: ExternalName
  {{- if $isIntercepted }}
  externalName: {{.Spec.Intercept.ToDevice}}.wg-{{$accountName}}.{{$clusterDnsSuffix}}
  {{- else}}
  externalName: {{.Name}}-internal.{{.Namespace}}.{{$clusterDnsSuffix}}
  {{- end }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{.Name}}-internal
  namespace: {{.Namespace}}
spec:
  selector:
    app: {{.Name}}
  ports:
    {{- range $svc := .Spec.Services }}
    {{- with $svc }}
    - protocol: {{.Type | upper | default "TCP"}}
      port: {{.Port}}
      name: {{.Port | squote}}
      targetPort: {{.TargetPort}}
    {{- end }}
    {{- end }}
{{- end}}

{{- if $isHpaEnabled }}
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{ $ownerRefs | toYAML | nindent 4}}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{.Name}}
  minReplicas: {{ .Spec.Hpa.MinReplicas }}
  maxReplicas: {{ .Spec.Hpa.MaxReplicas }}
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: {{.Spec.Hpa.ThresholdCpu}}
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: {{.Spec.Hpa.ThresholdMemory}}
{{- end }}
{{- end }}
