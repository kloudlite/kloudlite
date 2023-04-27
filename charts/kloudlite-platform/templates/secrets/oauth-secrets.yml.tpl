---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.secrets.names.oAuthSecret}}
  namespace: {{.Release.Namespace}}
data:
  github-app-pk.pem: |+
    {{.Values.oAuth2.github.appPrivateKey}}

stringData:
  GITHUB_CALLBACK_URL: {{printf .Values.oAuth2.github.callbackUrl .Values.routers.authWeb.domain }}
  GITHUB_WEBHOOK_URL: {{printf .Values.oAuth2.github.webhookUrl .Values.routers.webhooksApi.domain }}
  GITLAB_WEBHOOK_URL: {{printf .Values.oAuth2.gitlab.webhookUrl .Values.routers.webhooksApi.domain}}

  GITHUB_CLIENT_ID: {{.Values.oAuth2.github.clientId |squote}}
  GITHUB_CLIENT_SECRET: {{.Values.oAuth2.github.clientSecret}}
  GITHUB_APP_ID: {{.Values.oAuth2.github.appId | squote}}
  GITHUB_SCOPES: "user:email,admin:org"

  GITLAB_CALLBACK_URL: {{printf .Values.oAuth2.gitlab.callbackUrl .Values.routers.authWeb.domain }}
  GITLAB_CLIENT_ID: {{.Values.oAuth2.gitlab.clientId |squote}}
  GITLAB_CLIENT_SECRET: {{.Values.oAuth2.gitlab.clientSecret}}
  GITLAB_SCOPES: "api,read_repository"

  GOOGLE_CALLBACK_URL: {{printf .Values.oAuth2.google.callbackUrl .Values.routers.authWeb.domain }}
  GOOGLE_CLIENT_ID: {{.Values.oAuth2.google.clientId |squote}}
  GOOGLE_CLIENT_SECRET: {{.Values.oAuth2.google.clientSecret}}
  GOOGLE_SCOPES: "https://www.googleapis.com/auth/userinfo.profile,https://www.googleapis.com/auth/userinfo.email"
---
