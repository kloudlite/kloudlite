{{- $clusterIdentitySecret := (lookup "v1" "Secret" .Release.Namespace .Values.clusterIdentitySecretName) -}} 

{{ $x := (len $clusterIdentitySecret) }}

apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.clusterIdentitySecretName}}
  namespace: {{.Release.Namespace}}
stringData:
  CLUSTER_TOKEN: {{ .Values.clusterToken | default "" | squote }}
  {{- if (
    or 
      (eq $x 0) 
      (gt (len .Values.accessToken) 0) 
    ) 
  }}
  ACCESS_TOKEN: {{ .Values.accessToken | default "" | squote }}
  {{- end }}
