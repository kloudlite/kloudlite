{{- $namespace := get . "namespace" -}}
{{- $region := get . "region" -}}
{{- $localStorageClass := get . "local-storage-class" -}}
{{- $sharedConstants := get . "shared-constants" -}}
{{- $ownerRefs := get . "owner-refs" | default list -}}

{{- with $sharedConstants -}}
{{/*gotype: operators.kloudlite.io/apis/cluster-setup/v1.SharedConstants*/}}

apiVersion: crds.kloudlite.io/v1
kind: ManagedService
metadata:
  name: {{.MongoSvcName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
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
  name: {{.AuthRedisName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  inputs:
    keyPrefix: auth
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.RedisSvcName}}
---
# console-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.ConsoleRedisName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  inputs:
    keyPrefix: console
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.RedisSvcName}}

---
# ci-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.CiRedisName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  inputs:
    keyPrefix: ci
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.RedisSvcName}}
---
# dns-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.DnsRedisName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  inputs:
    keyPrefix: dns
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.RedisSvcName}}

---
# finance-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.FinanceRedisName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  inputs:
    keyPrefix: finance
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.RedisSvcName}}

---
# iam-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.IamRedisName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  inputs:
    keyPrefix: iam
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.RedisSvcName}}

---
# socket-redis
apiVersion: crds.kloudlite.io/v1
kind: ManagedResource
metadata:
  name: {{.SocketRedisName}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
spec:
  inputs:
    keyPrefix: socket
  mresKind:
    kind: ACLAccount
  msvcRef:
    apiVersion: redis.msvc.kloudlite.io/v1
    kind: StandaloneService
    name: {{.RedisSvcName}}
{{end}}
