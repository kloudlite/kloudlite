apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
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
      containers:
      {{- include "TemplateContainer" .Spec.Containers | indent 6 }}
