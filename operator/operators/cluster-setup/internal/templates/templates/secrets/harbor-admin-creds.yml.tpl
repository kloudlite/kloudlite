{{- $namespace := get . "namespace" -}}

apiVersion: v1
kind: Secret
metadata:
  name: harbor-admin-creds
  namespace: {{$namespace}}
stringData:
  HARBOR_ADMIN_USERNAME: $HARBOR_ADMIN_USERNAME
  HARBOR_ADMIN_PASSWORD: $HARBOR_ADMIN_PASSWORD
  HARBOR_IMAGE_REGISTRY_HOST: $HARBOR_REGISTRY_HOST
  HARBOR_API_VERSION: "v2.0"
  HARBOR_WEBHOOK_ENDPOINT: "$HARBOR_WEBHOOK_ENDPOINT"
  HARBOR_WEBHOOK_AUTHZ: "$HARBOR_WEBHOOK_AUTHZ"
  # DOCKER_SECRET_NAME: "kloudlite-docker-registry"

  DOCKER_SECRET_NAME: "kloudlite-harbor-creds"
  SERVICE_ACCOUNT_NAME: "kloudlite-svc-account"
  HARBOR_WEBHOOK_NAME: "kloudlite-docker-registry"
