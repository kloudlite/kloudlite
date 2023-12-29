---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.webhookSecrets.name}}
  namespace: {{.Release.Namespace}}
stringData:
  GITHUB_AUTHZ_SECRET:  {{.Values.webhookSecrets.githubSecret}}
  GITLAB_AUTHZ_SECRET: {{.Values.webhookSecrets.gitlabSecret}}
  KLOUDLITE_AUTHZ_SECRET: {{.Values.webhookSecrets.kloudliteSecret}}
---
