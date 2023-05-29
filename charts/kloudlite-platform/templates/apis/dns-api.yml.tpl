apiVersion: crds.kloudlite.io/v1
kind: App
metadata:
  name: {{.Values.apps.dnsApi.name}}
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
    - port: 3001
      targetPort: 3001
      name: grpc
      type: tcp
  containers:
    - name: main
      image: {{.Values.apps.dnsApi.image}}
      imagePullPolicy: {{.Values.apps.authApi.ImagePullPolicy | default .Values.imagePullPolicy }}

      resourceCpu:
        min: "40m"
        max: "80m"
      resourceMemory:
        min: "60Mi"
        max: "100Mi"

      env:
        - key: DNS_DOMAIN_NAMES
          value: {{ .Values.apps.dnsApi.configuration.dnsNames | join "," }}

        - key: EDGE_CNAME_BASE_DOMAIN
          value: "{{ .Values.apps.dnsApi.configuration.edgeCNAME }}"

        - key: REDIS_HOSTS
          type: secret
          refName: mres-{{.Values.managedResources.dnsRedis}}
          refKey: HOSTS

        - key: REDIS_PASSWORD
          type: secret
          refName: mres-{{.Values.managedResources.dnsRedis}}
          refKey: PASSWORD

        - key: REDIS_PREFIX
          type: secret
          refName: mres-{{.Values.managedResources.dnsRedis}}
          refKey: PREFIX

        - key: REDIS_USERNAME
          type: secret
          refName: mres-{{.Values.managedResources.dnsRedis}}
          refKey: USERNAME

        - key: MONGO_URI
          type: secret
          refName: mres-{{.Values.managedResources.dnsDb}}
          refKey: URI

        - key: MONGO_DB_NAME
          value: {{.Values.managedResources.dnsDb}}

        - key: CONSOLE_SERVICE
          value: "{{.Values.apps.consoleApi}}.{{.Release.Namespace}}.svc.cluster.local:3001"

        - key: FINANCE_SERVICE
          value: "{{.Values.apps.financeApi}}.{{.Release.Namespace}}.svc.cluster.local:3001"

        - key: PORT
          value: '3000'

        - key: GRPC_PORT
          value: '3001'

        - key: DNS_PORT
          value: '5353'

---

apiVersion: v1
kind: Service
metadata:
  name: {{.Values.apps.dnsApi.name}}-exposed
  namespace: {{.Release.Namespace}}
spec:
  ports:
    - name: server-udp
      nodePort: 30053
      port: 5353
      protocol: UDP
      targetPort: 5353
  selector:
    app: {{.Values.apps.dnsApi.name}}
  type: NodePort
---
