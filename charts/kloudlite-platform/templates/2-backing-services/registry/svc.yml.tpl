{{- $name := .Release.Name -}}
{{- $namespace := .Release.Namespace -}}

apiVersion: v1
kind: Service
metadata:
  name: distribution
spec:
  ports:
  - name: "registry"
    port: 80
    protocol: TCP
    targetPort: http
  selector:
    app: {{ $name }}
