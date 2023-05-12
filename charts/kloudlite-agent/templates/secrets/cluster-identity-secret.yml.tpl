{{- $clusterIdentitySecret := (lookup "v1" "Secret" .Release.Namespace .Values.clusterIdentitySecretName) -}} 

{{ $x := (len $clusterIdentitySecret) }}

{{- if not (eq $x 0) }}
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.clusterIdentitySecretName}}
  namespace: {{.Release.Namespace}}
stringData:
  CLUSTER_TOKEN: {{.Values.clusterToken | default ""}}
  ACCESS_TOKEN: {{.Values.accessToken | default ""}}
{{- end }}
