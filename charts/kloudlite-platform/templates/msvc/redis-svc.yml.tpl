apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{.Values.managedServices.redisSvc}}
  namespace: {{.Release.Namespace}}
  labels:
    kloudlite.io/account-ref: {{.Values.accountName}}
spec:
  {{/* {{- if .Values.region}} */}}
  {{/* region: {{.Values.region}} */}}
  {{/* {{- end }} */}}
  msvcKind:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
  inputs:
    resources:
      cpu:
        min: 200m
        max: 300m
      memory: 300Mi
      storage:
        size: 1Gi
        {{- if .Values.persistence.storageClassName }}
        storageClass: {{.Values.persistence.storageClassName}}
        {{- end}}
