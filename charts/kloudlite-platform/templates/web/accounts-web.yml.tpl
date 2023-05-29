apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.accountsWeb.name}}
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
      image: {{.Values.apps.accountsWeb.name}}
      imagePullPolicy: {{.Values.apps.authWeb.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "100m"
        max: "200m"
      resourceMemory:
        min: "200Mi"
        max: "300Mi"
      env:
        - key: BASE_URL
          value: {{.Values.baseDomain}}
        - key: PORT
          value: "3000"
---
