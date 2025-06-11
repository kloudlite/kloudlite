{{- $secret := (lookup "v1" "Secret" .Release.Namespace (include "apps.webhooksApi.authenticationSecret.name" .)) -}}

---
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "apps.webhooksApi.authenticationSecret.name" . }}
  namespace: {{.Release.Namespace}}
stringData:
  GITHUB_AUTHZ_SECRET:  {{.Values.apps.webhooksApi.authenticationSecrets.githubAuthzSecret}}
  GITLAB_AUTHZ_SECRET: {{.Values.apps.webhooksApi.authenticationSecrets.gitlabAuthzSecret}}
  KLOUDLITE_AUTHZ_SECRET: {{.Values.apps.webhooksApi.authenticationSecrets.kloudliteAuthzSecret}}
data:
  {{include "apps.webhooksApi.authenticationSecret.token-key" .}}: |+
    {{- if $secret }}
    {{ dig "data" (include "apps.webhooksApi.authenticationSecret.token-key" .) "." $secret }}
    {{- else }}
    {{ randBytes 64 | sha256sum | b64enc }}
    {{- end }}
---
