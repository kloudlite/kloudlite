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
---
