{{- $namespace := get . "namespace" -}}
apiVersion: v1
kind: Secret
metadata:
  name: project-operator-env
  namespace: {{$namespace}}
stringData:
  PROJECT_CONFIGMAP_NAME: project-config
  DOCKER_SECRET_NAME: kloudlite-docker-registry
  ADMIN_ROLE_NAME: kloudlite-ns-admin
  SVC_ACCOUNT_NAME: kloudlite-svc-account
  ACCOUNT_ROUTER_NAME: account-router
