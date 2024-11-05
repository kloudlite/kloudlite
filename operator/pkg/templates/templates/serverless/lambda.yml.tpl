{{- $ownerRefs := get . "owner-refs"}}
{{- $obj := get . "object" }}
{{- $volumes := get . "volumes"}}
{{- $vMounts := get . "volumeMounts"}}

{{- with $obj }}
{{- /* gotype: github.com/kloudlite/operator/apis/serverless/v1.Lambda*/ -}}
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
  {{- if .Labels}}
  labels: {{.Labels | toYAML | nindent 4}}
  {{- end }}
spec:
  template:
    metadata:
      deletionGracePeriodSeconds: 5
      annotations:
        autoscaling.knative.dev/class: kpa.autoscaling.knative.dev
        autoscaling.knative.dev/metric: rps
        autoscaling.knative.dev/target: "{{.Spec.TargetRps}}"
        autoscaling.knative.dev/min-scale: "{{.Spec.MinScale}}"
        autoscaling.knative.dev/max-scale: "{{.Spec.MaxScale}}"
      {{- if .Labels}}
      labels: {{.Labels | toYAML | nindent 8}}
      {{- end }}
    spec:
      {{- if .Spec.NodeSelector }}
      nodeSelector: {{.Spec.NodeSelector | toYAML | nindent 8}}
      {{- end }}

      {{- if .Spec.Tolerations }}
      tolerations: {{.Spec.Tolerations | toYAML | nindent 8}}
      {{- end}}
      serviceAccountName: {{.Spec.ServiceAccount}}
      {{- $myDict := dict "containers" .Spec.Containers "volumeMounts" $vMounts }}
      containers: {{- include "TemplateContainer" $myDict | nindent 8 }}
      {{- if $volumes }}
      volumes: {{- $volumes | toYAML | nindent 8 }}
      {{- end}}
{{- end }}
