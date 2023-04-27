---
apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.secrets.names.stripeSecret}}
  namespace: {{.Release.Namespace}}
stringData:
  PUBLIC_KEY: {{.Values.stripe.publicKey}}
  SECRET_KEY: {{.Values.stripe.secretKey}}
---
