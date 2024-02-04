{{- $name :=  get . "name" }}
{{- $namespace :=  get . "namespace" }}
{{- $mongodbRootPassword :=  get . "mongodb-root-password" }}
{{- $mongodbReplicaSetKey := get . "mongodb-replica-set-key" }}

apiVersion: v1
kind: Secret
metadata:
	name: "{{$name}}"
	namespace: "{{$namespace}}"
stringData:
  mongodb-root-password: "{{$mongodbRootPassword}}"
  mongodb-replica-set-key: "{{$mongodbReplicaSetKey}}"
