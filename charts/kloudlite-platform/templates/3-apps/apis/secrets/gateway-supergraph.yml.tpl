apiVersion: v1
kind: ConfigMap
metadata:
  name: gateway-supergraph
  namespace: {{.Release.Namespace}}
data:
  config: |+
    serviceList:
      - name: auth-api
        url: http://auth-api/query
      - name: accounts-api
        url: http://accounts-api/query

      - name: container-registry-api
        url: http://container-registry-api/query

      - name: console-api
        url: http://console-api/query
      
      - name: infra-api
        url: http://infra-api/query


---
