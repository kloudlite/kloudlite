---
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "apps.authApi.oAuth2-secret.name" .}}
  namespace: {{.Release.Namespace}}
{{- if .Values.apps.authApi.oAuth2.providers.github.enabled }}
data:
  github-app-pk.pem: |+
   {{.Values.apps.authApi.oAuth2.providers.github.appPrivateKey}}
{{- end }}

stringData:
  OAUTH2_ENABLED: {{.Values.apps.authApi.oAuth2.enabled | squote }}
  OAUTH2_GITHUB_ENABLED: {{.Values.apps.authApi.oAuth2.providers.github.enabled | squote }}
  OAUTH2_GITLAB_ENABLED: {{.Values.apps.authApi.oAuth2.providers.gitlab.enabled | squote }}
  OAUTH2_GOOGLE_ENABLED: {{.Values.apps.authApi.oAuth2.providers.github.enabled | squote}}

  OAUTH2_GITHUB_ENABLED: {{.Values.apps.authApi.oAuth2.providers.github.enabled | squote}}
  {{- if .Values.apps.authApi.oAuth2.providers.github.enabled }}
  GITHUB_CALLBACK_URL: {{ .Values.apps.authApi.oAuth2.providers.github.callbackURL | squote }}
  GITHUB_CLIENT_ID: {{.Values.apps.authApi.oAuth2.providers.github.clientID |squote}}
  GITHUB_CLIENT_SECRET: {{.Values.apps.authApi.oAuth2.providers.github.clientSecret| squote}}
  GITHUB_APP_ID: {{.Values.apps.authApi.oAuth2.providers.github.appID | squote}}
  GITHUB_APP_NAME: {{.Values.apps.authApi.oAuth2.providers.github.githubAppName | squote}}
  GITHUB_SCOPES: "user:email,admin:org"
  {{- end }}

  OAUTH2_GITLAB_ENABLED: {{.Values.apps.authApi.oAuth2.providers.gitlab.enabled | squote}}
  {{- if .Values.apps.authApi.oAuth2.providers.gitlab.enabled }}
  GITLAB_CALLBACK_URL: {{ .Values.apps.authApi.oAuth2.providers.gitlab.callbackURL |squote }}
  GITLAB_CLIENT_ID: {{.Values.apps.authApi.oAuth2.providers.gitlab.clientID | squote}}
  GITLAB_CLIENT_SECRET: {{.Values.apps.authApi.oAuth2.providers.gitlab.clientSecret | squote}}
  GITLAB_SCOPES: "api,read_repository"
  {{- end }}

  OAUTH2_GOOGLE_ENABLED: {{.Values.apps.authApi.oAuth2.providers.google.enabled | squote}}
  {{- if .Values.apps.authApi.oAuth2.providers.google.enabled }}
  GOOGLE_CALLBACK_URL: {{ .Values.apps.authApi.oAuth2.providers.google.callbackURL | squote }}
  GOOGLE_CLIENT_ID: {{.Values.apps.authApi.oAuth2.providers.google.clientID | squote}}
  GOOGLE_CLIENT_SECRET: {{.Values.apps.authApi.oAuth2.providers.google.clientSecret | squote }}
  GOOGLE_SCOPES: "https://www.googleapis.com/auth/userinfo.profile,https://www.googleapis.com/auth/userinfo.email"
  {{- end }}
---
