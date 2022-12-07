{{- $namespace := get . "namespace" -}}
{{- $region := get . "region" -}}
{{- $localStorageClass := get . "local-storage-class" -}}

{{- $redisSvcName := "redis-svc" -}}

apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{$redisSvcName}}
  namespace: {{$namespace}}
spec:
  region: {{$region}}
  msvcKind:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
  inputs:
    resources:
      cpu:
        min: 200m
        max: 300m
      memory: 300Mi
    storage:
      size: 1Gi
      {{if $localStorageClass}}
      storageClass: {{$localStorageClass}}
      {{end}}

---
# auth-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: auth-redis
  namespace: {{$namespace}}
spec:
  inputs:
    keyPrefix: auth
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$redisSvcName}}
---
# console-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: console-redis
  namespace: {{$namespace}}
spec:
  inputs:
    keyPrefix: console
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$redisSvcName}}

---
# ci-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: ci-redis
  namespace: {{$namespace}}
spec:
  inputs:
    keyPrefix: ci
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$redisSvcName}}
---
# dns-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: dns-redis
  namespace: {{$namespace}}
spec:
  inputs:
    keyPrefix: dns
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$redisSvcName}}

---
# finance-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: finance-redis
  namespace: {{$namespace}}
spec:
  inputs:
    keyPrefix: finance
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$redisSvcName}}

---
# iam-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: iam-redis
  namespace: {{$namespace}}
spec:
  inputs:
    keyPrefix: iam
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$redisSvcName}}

---
# socket-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: socket-redis
  namespace: {{$namespace}}
spec:
  inputs:
    keyPrefix: socket
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{$redisSvcName}}
