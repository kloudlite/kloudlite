fullnameOverride: redpanda-operator

resources:
  limits:
    cpu: 60m
    memory: 120Mi
  requests:
    cpu: 20m
    memory: 80Mi

webhook:
  enabled: false
