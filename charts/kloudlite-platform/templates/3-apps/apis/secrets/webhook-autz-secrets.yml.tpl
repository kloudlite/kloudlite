---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.webhookSecrets.name}}
  namespace: {{.Release.Namespace}}
stringData:
  GITHUB_SECRET: {{.Values.webhookSecrets.githubSecret}}
  GITLAB_SECRET: {{.Values.webhookSecrets.gitlabSecret}}
  KLOUDLITE_SECRET: {{.Values.webhookSecrets.kloudliteSecret}}
---
