{{- $obj := get . "obj" }}
{{- $volumes := get . "volumes"}}
{{- $vMounts := get . "volumeMounts"}}
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: {{$obj.Name}}
  namespace: {{$obj.Namespace}}
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/class: kpa.autoscaling.knative.dev
        autoscaling.knative.dev/metric: concurrency
        autoscaling.knative.dev/target: "10"
        autoscaling.knative.dev/min-scale: "0"
        # Limit scaling to 100 pods.
        autoscaling.knative.dev/max-scale: "100"
    spec:
      serviceAccountName: kloudlite-svc-account
      containers:
      {{- $myDict := dict "containers" $obj.Spec.Containers "volumeMounts" $vMounts }}
      {{- include "TemplateContainer" $myDict | indent 6 }}
      {{- if $volumes }}
      volumes: {{- $volumes| toPrettyJson | indent 6 }}
      {{- end}}
