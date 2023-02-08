{{ $caData := get . "ca-data"}}
{{ $userToken := get . "user-token"}}
{{ $clusterEndpoint := get . "cluster-endpoint"}}
{{- $userName := get . "user-name" -}}

apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {{$caData}}
    server: {{$clusterEndpoint}}
  name: kloudlite
contexts:
- context:
    cluster: kloudlite
    namespace: kube-system
    user: {{$userName}}
  name: {{$userName}}
current-context: {{$userName}}
kind: Config
preferences: {}
users:
- name: {{$userName}}
  user:
    token: {{$userToken}}
