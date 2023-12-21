---
apiVersion: v1
kind: Secret
metadata:
  name: kloudlite-cloudflare
  namespace: {{.Release.Namespace}}
stringData:
    api_token: {{.Values.cloudflareWildCardCert.cloudflareCreds.secretToken}}