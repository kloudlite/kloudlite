apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{.Values.managedServices.mongoSvc}}
  namespace: {{.Release.Namespace}}
  labels:
    
spec:
  {{/* {{- if .Values.region}} */}}
  {{/* region: {{.Values.region}} */}}
  {{/* {{- end }} */}}
  msvcKind:
    apiVersion: mongodb.msvc.kloudlite.io/v1
    kind: StandaloneService
  inputs:
    resources:
      cpu:
        min: 400m
        max: 500m
      memory: 500Mi
      storage:
        size: 1Gi
        {{- if .Values.persistence.storageClassName }}
        storageClass: {{.Values.persistence.storageClassName}}
        {{- end}}
