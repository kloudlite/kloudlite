apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.socketWeb.name}}
  namespace: {{.Release.Namespace}}
  labels:
    
spec:
  region: {{.Values.region | default ""}}
  serviceAccount: {{.Values.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}

  services:
    - port: 80
      targetPort: 3000
      name: socket
      type: tcp
    - port: 3001
      targetPort: 3001
      name: http
      type: tcp

  containers:
    - name: main
      image: {{.Values.apps.socketWeb.name}}
      imagePullPolicy: {{.Values.apps.socketWeb.ImagePullPolicy | default .Values.imagePullPolicy }}
      resourceCpu:
        min: "60m"
        max: "100m"
      resourceMemory:
        min: "100Mi"
        max: "140Mi"
      env:
        - key: BASE_URL
          value: {{.Values.baseDomain}}
        - key: REDIS_URI
          type: secret
          refName: mres-{{.Values.managedResources.socketWebRedis}}
          refKey: URI
---
