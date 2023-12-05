{{- $name :=  get . "name" }}
{{- $namespace :=  get . "namespace" }}
{{- $mongodbRootPassword :=  get . "mongodb-root-password" }}

apiVersion: v1
kind: Secret
metadata:
	name: "{{$name}}"
	namespace: "{{$namespace}}"
stringData:
  mongodb-root-password: "{{$mongodbRootPassword}}"
