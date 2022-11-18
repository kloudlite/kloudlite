{{- $name := get . "name"  -}}
{{- $namespace := get . "namespace"  -}}
{{- $nodeSelector := get . "node-selector" | default dict -}}
{{/*{{- $tolerations := get . "tolerations" | default list -}}*/}}
{{- $ownerRefs := get . "owner-refs" -}}
{{- $doSecretName := get . "do-secret-name" -}}
{{- $doSecretKey := get . "do-secret-key" -}}

apiVersion: csi.helm.kloudlite.io/v1
kind: DigitaloceanCSIDriver
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4 }}
spec:
  name: {{$name}}
  namespace: {{$namespace}}
  serviceAccountName: "kloudlite-cluster-svc-account"
  digitalocean:
    accessToken:
      secretName: {{$doSecretName}}
      secretKey: {{$doSecretKey}}

  controller:
    nodeSelector: {{$nodeSelector | toYAML | nindent 6 }}
    tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists

  daemonset:
    podLabels: {}
    podAnnotations: {}
    nodeSelector: {{$nodeSelector | toYAML | nindent 6 }}
    tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
