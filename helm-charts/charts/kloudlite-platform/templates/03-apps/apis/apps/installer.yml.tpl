{{- if .Values.apps.klInstaller.install }}

{{- $appName := include "apps.klInstaller.name" . }}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{ $appName | squote }}
  namespace: {{.Release.Namespace}}
  labels: {{ include "common.pod-labels" . | nindent 4}}
  annotations:
    {{ include "common.pod-annotations" . | nindent 4}}

spec:
  serviceAccount: {{.Values.serviceAccounts.normal.name}}

  nodeSelector: {{ .Values.scheduling.stateless.nodeSelector | toYaml | nindent 4 }}
  tolerations: {{ .Values.scheduling.stateless.tolerations | toYaml | nindent 4 }}
  
  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  services:
    - port: {{ include "apps.klInstaller.httpPort" . }}

  {{- /* hpa: */}}
  {{- /*   enabled: false */}}
  {{- /*   minReplicas: {{.Values.apps.klInstaller.minReplicas}} */}}
  {{- /*   maxReplicas: {{.Values.apps.klInstaller.maxReplicas}} */}}
  {{- /*   thresholdCpu: 70 */}}
  {{- /*   thresholdMemory: 80 */}}

  containers:
    - name: main
      image: {{.Values.apps.klInstaller.image.repository}}:{{.Values.apps.klInstaller.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}

{{- end }}
