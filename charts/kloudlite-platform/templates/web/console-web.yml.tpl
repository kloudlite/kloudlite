apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.consoleWeb.name}}
  namespace: {{.Release.Namespace}}
spec:
  region: {{.Values.region | default ""}}
  serviceAccount: {{.Values.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.consoleWeb.image}}
      imagePullPolicy: {{.Values.apps.consoleWeb.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "200Mi"
        max: "300Mi"
      livenessProbe: &probe
        type: httpGet
        initialDelay: 5
        failureThreshold: 3
        httpGet:
          path: /console/assets/healthy.txt
          port: 3000
        interval: 10
      readinessProbe: *probe
      env:
        - key: BASE_URL
          value: {{.Values.baseDomain}}
        - key: PORT
          value: "3000"
        - key: GITHUB_APP
          value: "{{.Values.githubAppName}}"
---
