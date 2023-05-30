{{- if .Values.apps.authApi.configuration.oAuth2.enabled }}

---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.secretNames.oAuthSecret}}
  namespace: {{.Release.Namespace}}
data:
  {{- if .Values.apps.authApi.configuration.oAuth2.github.enabled }}
  github-app-pk.pem: |+
    {{.Values.apps.authApi.configuration.oAuth2.github.appPrivateKey}}
  {{- end }}

stringData:
  OAUTH2_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.enabled}}
  OAUTH2_GITHUB_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.github.enabled}}
  OAUTH2_GITLAB_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.gitlab.enabled}}
  OAUTH2_GOOGLE_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.github.enabled}}

  {{- if .Values.apps.authApi.configuration.oAuth2.github.enabled }}
  GITHUB_CALLBACK_URL: {{ .Values.apps.authApi.configuration.oAuth2.github.callbackUrl }}
  {{/* GITHUB_WEBHOOK_URL: {{ .Values.apps.authApi.configuration.oAuth2.github.webhookUrl }} */}}
  {{/* GITLAB_WEBHOOK_URL: {{ .Values.apps.authApi.configuration.oAuth2.gitlab.webhookUrl }} */}}

  GITHUB_CLIENT_ID: {{.Values.apps.authApi.configuration.oAuth2.github.clientId |squote}}
  GITHUB_CLIENT_SECRET: {{.Values.apps.authApi.configuration.oAuth2.github.clientSecret}}
  GITHUB_APP_ID: {{.Values.apps.authApi.configuration.oAuth2.github.appId | squote}}
  GITHUB_SCOPES: "user:email,admin:org"
  {{- end }}

  {{- if .Values.apps.authApi.configuration.oAuth2.gitlab.enabled }}
  GITLAB_CALLBACK_URL: {{ .Values.apps.authApi.configuration.oAuth2.gitlab.callbackUrl }}
  GITLAB_CLIENT_ID: {{.Values.apps.authApi.configuration.oAuth2.gitlab.clientId |squote}}
  GITLAB_CLIENT_SECRET: {{.Values.apps.authApi.configuration.oAuth2.gitlab.clientSecret}}
  GITLAB_SCOPES: "api,read_repository"
  {{- end }}

  {{- if .Values.apps.authApi.configuration.oAuth2.google.enabled }}
  GOOGLE_CALLBACK_URL: {{ .Values.apps.authApi.configuration.oAuth2.google.callbackUrl }}
  GOOGLE_CLIENT_ID: {{.Values.apps.authApi.configuration.oAuth2.google.clientId |squote}}
  GOOGLE_CLIENT_SECRET: {{.Values.apps.authApi.configuration.oAuth2.google.clientSecret}}
  GOOGLE_SCOPES: "https://www.googleapis.com/auth/userinfo.profile,https://www.googleapis.com/auth/userinfo.email"
  {{- end }}
---

{{- end }}
