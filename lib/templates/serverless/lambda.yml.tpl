{{- $ownerRefs := get . "owner-refs"}}
{{- $obj := get . "object" }}
{{- $volumes := get . "volumes"}}
{{- $vMounts := get . "volumeMounts"}}

{{- with $obj }}
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  template:
    metadata:
      deletionGracePeriodSeconds: 5
      annotations:
        autoscaling.knative.dev/class: kpa.autoscaling.knative.dev
        autoscaling.knative.dev/metric: rps
        autoscaling.knative.dev/target: "100"
        autoscaling.knative.dev/min-scale: "1"
        # Limit scaling to 100 pods.
        autoscaling.knative.dev/max-scale: "100"
    spec:
      serviceAccountName: kloudlite-svc-account
      {{- $myDict := dict "containers" .Spec.Containers "volumeMounts" $vMounts }}
      containers: {{- include "TemplateContainer" $myDict | nindent 8 }}
      {{- if $volumes }}
      volumes: {{- $volumes | toYAML | nindent 8 }}
      {{- end}}
{{- end }}
