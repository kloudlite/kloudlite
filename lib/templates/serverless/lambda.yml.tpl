{{- $obj := get . "obj" }}
{{- $volumes := get . "volumes"}}
{{- $vMounts := get . "volumeMounts"}}

{{- with $obj }}
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  template:
    metadata:
      deletionGracePeriodSeconds: 5
      annotations:
        autoscaling.knative.dev/class: kpa.autoscaling.knative.dev
        autoscaling.knative.dev/metric: rps
        autoscaling.knative.dev/target: "100"
        autoscaling.knative.dev/min-scale: "3"
        # Limit scaling to 100 pods.
        autoscaling.knative.dev/max-scale: "100"
    spec:
      serviceAccountName: kloudlite-svc-account
      containers:
      {{- $myDict := dict "containers" .Spec.Containers "volumeMounts" $vMounts }}
      {{- include "TemplateContainer" $myDict | indent 6 }}
      {{- if $volumes }}
      volumes: {{- $volumes| toYAML | indent 6 }}
      {{- end}}
{{- end }}
