{{- if .Values.apps.kloudliteWebsite.enabled -}}
apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: kloudlite-website
  namespace: {{.Release.Namespace}}
spec:
  serviceAccount: {{.Values.global.normalSvcAccount}}

  {{ include "node-selector-and-tolerations" . | nindent 2 }}
  
  services:
    - port: 80
      targetPort: 3000
      name: http
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.kloudliteWebsite.image}}
      imagePullPolicy: {{.Values.global.imagePullPolicy }}
      resourceCpu:
        min: "100m"
        max: "100m"
      resourceMemory:
        min: "100Mi"
        max: "100Mi"
      {{- /* livenessProbe: &probe */}}
      {{- /*   type: httpGet */}}
      {{- /*   initialDelay: 5 */}}
      {{- /*   failureThreshold: 3 */}}
      {{- /*   httpGet: */}}
      {{- /*     path: /healthy.txt */}}
      {{- /*     port: 3000 */}}
      {{- /*   interval: 10 */}}
      {{- /* readinessProbe: *probe */}}
      env:
        - key: PORT
          value: "3000"
---
{{- end -}}

