apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
spec:
  template:
    spec:
      containers:
      {{- include "TemplateContainer" .Spec.Containers | indent 6}}
