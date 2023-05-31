---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.secretNames.oAuthSecret}}
  namespace: {{.Release.Namespace}}
{{- if .Values.apps.authApi.configuration.oAuth2.github.enabled }}
data:
  github-app-pk.pem: |+
    {{.Values.apps.authApi.configuration.oAuth2.github.appPrivateKey}}
{{- end }}

stringData:
  OAUTH2_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.enabled | squote }}
  OAUTH2_GITHUB_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.github.enabled | squote }}
  OAUTH2_GITLAB_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.gitlab.enabled | squote }}
  OAUTH2_GOOGLE_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.github.enabled | squote}}

  OAUTH2_GITHUB_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.github.enabled | squote}}
  {{- if .Values.apps.authApi.configuration.oAuth2.github.enabled }}
  GITHUB_CALLBACK_URL: {{ .Values.apps.authApi.configuration.oAuth2.github.callbackUrl | squote }}
  GITHUB_CLIENT_ID: {{.Values.apps.authApi.configuration.oAuth2.github.clientId |squote}}
  GITHUB_CLIENT_SECRET: {{.Values.apps.authApi.configuration.oAuth2.github.clientSecret| squote}}
  GITHUB_APP_ID: {{.Values.apps.authApi.configuration.oAuth2.github.appId | squote}}
  GITHUB_SCOPES: "user:email,admin:org"
  {{- end }}

  OAUTH2_GITLAB_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.gitlab.enabled | squote}}
  {{- if .Values.apps.authApi.configuration.oAuth2.gitlab.enabled }}
  GITLAB_CALLBACK_URL: {{ .Values.apps.authApi.configuration.oAuth2.gitlab.callbackUrl |squote }}
  GITLAB_CLIENT_ID: {{.Values.apps.authApi.configuration.oAuth2.gitlab.clientId | squote}}
  GITLAB_CLIENT_SECRET: {{.Values.apps.authApi.configuration.oAuth2.gitlab.clientSecret | squote}}
  GITLAB_SCOPES: "api,read_repository"
  {{- end }}

  OAUTH2_GOOGLE_ENABLED: {{.Values.apps.authApi.configuration.oAuth2.google.enabled | squote}}
  {{- if .Values.apps.authApi.configuration.oAuth2.google.enabled }}
  GOOGLE_CALLBACK_URL: {{ .Values.apps.authApi.configuration.oAuth2.google.callbackUrl | squote }}
  GOOGLE_CLIENT_ID: {{.Values.apps.authApi.configuration.oAuth2.google.clientId | squote}}
  GOOGLE_CLIENT_SECRET: {{.Values.apps.authApi.configuration.oAuth2.google.clientSecret | squote }}
  GOOGLE_SCOPES: "https://www.googleapis.com/auth/userinfo.profile,https://www.googleapis.com/auth/userinfo.email"
  {{- end }}
---
