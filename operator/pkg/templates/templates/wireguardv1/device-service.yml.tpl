kind: Service
apiVersion: v1
metadata:
  name: {{.Name}}
  namespace: {{.Namespace}}
  labels: {{.Labels | toJson }}
  annotations: {{.Annotations | toJson }}
  ownerReferences: {{.OwnerReferences | toJson}}
spec:
  ports: {{.Spec.Ports | toJson}}
  selector: {{.Spec.Selector | toJson}}
