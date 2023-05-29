---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.secretNames.webhookAuthzSecret}}
  namespace: {{.Release.Namespace}}
stringData:
  GITHUB_SECRET: {{.Values.apps.webhooksApi.configuration.webhookAuthz.githubSecret}}
  GITLAB_SECRET: {{.Values.apps.webhooksApi.configuration.webhookAuthz.gitlabSecret}}
  HARBOR_SECRET: {{.Values.apps.webhooksApi.configuration.webhookAuthz.harborSecret}}
  KLOUDLITE_SECRET: {{.Values.apps.webhooksApi.configuration.webhookAuthz.kloudliteSecret}}
---
