---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.oAuth.secretName}}
  namespace: {{.Release.Namespace}}
{{- if .Values.oAuth.providers.github.enabled }}
data:
  github-app-pk.pem: |+
   {{.Values.oAuth.providers.github.appPrivateKey}}
{{- end }}

stringData:
  OAUTH2_ENABLED: {{.Values.oAuth.enabled | squote }}
  OAUTH2_GITHUB_ENABLED: {{.Values.oAuth.providers.github.enabled | squote }}
  OAUTH2_GITLAB_ENABLED: {{.Values.oAuth.providers.gitlab.enabled | squote }}
  OAUTH2_GOOGLE_ENABLED: {{.Values.oAuth.providers.github.enabled | squote}}

  OAUTH2_GITHUB_ENABLED: {{.Values.oAuth.providers.github.enabled | squote}}
  {{- if .Values.oAuth.providers.github.enabled }}
  GITHUB_CALLBACK_URL: {{ .Values.oAuth.providers.github.callbackUrl | squote }}
  GITHUB_CLIENT_ID: {{.Values.oAuth.providers.github.clientId |squote}}
  GITHUB_CLIENT_SECRET: {{.Values.oAuth.providers.github.clientSecret| squote}}
  GITHUB_APP_ID: {{.Values.oAuth.providers.github.appId | squote}}
  GITHUB_SCOPES: "user:email,admin:org"
  {{- end }}

  OAUTH2_GITLAB_ENABLED: {{.Values.oAuth.providers.gitlab.enabled | squote}}
  {{- if .Values.oAuth.providers.gitlab.enabled }}
  GITLAB_CALLBACK_URL: {{ .Values.oAuth.providers.gitlab.callbackUrl |squote }}
  GITLAB_CLIENT_ID: {{.Values.oAuth.providers.gitlab.clientId | squote}}
  GITLAB_CLIENT_SECRET: {{.Values.oAuth.providers.gitlab.clientSecret | squote}}
  GITLAB_SCOPES: "api,read_repository"
  {{- end }}

  OAUTH2_GOOGLE_ENABLED: {{.Values.oAuth.providers.google.enabled | squote}}
  {{- if .Values.oAuth.providers.google.enabled }}
  GOOGLE_CALLBACK_URL: {{ .Values.oAuth.providers.google.callbackUrl | squote }}
  GOOGLE_CLIENT_ID: {{.Values.oAuth.providers.google.clientId | squote}}
  GOOGLE_CLIENT_SECRET: {{.Values.oAuth.providers.google.clientSecret | squote }}
  GOOGLE_SCOPES: "https://www.googleapis.com/auth/userinfo.profile,https://www.googleapis.com/auth/userinfo.email"
  {{- end }}
---
