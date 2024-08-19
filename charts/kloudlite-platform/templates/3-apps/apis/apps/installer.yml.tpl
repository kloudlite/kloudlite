{{- $appName := "installer" }}

apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: kl-installer
  namespace: {{.Release.Namespace}}
spec:
  {{ include "node-selector-and-tolerations" . | nindent 2 }}
  tolerations: {{.Values.nodepools.stateless.tolerations | toYaml | nindent 4}}
  nodeSelector: {{.Values.nodepools.stateless.labels | toYaml | nindent 4}}

  topologySpreadConstraints:
    {{ include "tsc-hostname" (dict "kloudlite.io/app.name" $appName) | nindent 4 }}

  services:
    - port: 3000

  hpa:
    enabled: true
    minReplicas: {{.Values.apps.klInstaller.minReplicas}}
    maxReplicas: {{.Values.apps.klInstaller.maxReplicas}}
    thresholdCpu: 70
    thresholdMemory: 80

  containers:
    - name: main
      image: {{.Values.apps.klInstaller.image.repository}}:{{.Values.apps.klInstaller.image.tag | default (include "image-tag" .) }}
      imagePullPolicy: {{ include "image-pull-policy" .}}
      {{- /* env: */}}
      {{- /*   - key: DEFAULT_USER */}}
      {{- /*     value: "kloudlite" */}}
      {{- /**/}}
      {{- /*   - key: FORCE_USER */}}
      {{- /*     value: "kloudlite" */}}
