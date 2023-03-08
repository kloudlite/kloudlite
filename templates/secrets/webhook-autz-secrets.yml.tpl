---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.secrets.names.webhookAuthzSecret}}
  namespace: {{.Release.Namespace}}
stringData:
  GITHUB_SECRET: {{.Values.webhookAuthz.githubSecret}}
  GITLAB_SECRET: {{.Values.webhookAuthz.gitlabSecret}}
  HARBOR_SECRET: {{.Values.webhookAuthz.harborSecret}}
  KLOUDLITE_SECRET: {{.Values.webhookAuthz.kloudliteSecret}}
---
