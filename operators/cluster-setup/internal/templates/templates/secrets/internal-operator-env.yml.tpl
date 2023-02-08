{{- $namespace := get . "namespace" -}}
{{- $podCidr := get . "pod-cidr" | default "10.42.0.0/16" -}}
{{- $svcCidr := get . "svc-cidr" | default "10.43.0.0/16" -}}
{{- $clusterId := get . "cluster-id" -}}
{{- $wildcardDomain := get . "wildcard-domain" -}}

{{- $nameserverEndpoint := get . "nameserver-endpoint" -}}
{{- $nameserverUsername := get . "nameserver-username" -}} 
{{- $nameserverPassword := get . "nameserver-password" -}} 

apiVersion: v1
kind: Secret
metadata:
  name: internal-operator-env
  namespace: {{$namespace}}
stringData:
  COMM: "true"
  INFRA: "true"
  WG_DOMAIN: '{{$wildcardDomain}}'

  SSH_PATH: /home/nonroot/ssh
  STORE_PATH: /terraform/storage
  TF_TEMPLATES_PATH: /templates/terraform

{{/*  NAMESERVER_USER: kloudlite-dns-admin*/}}
{{/*  NAMESERVER_PASSWORD: 'rX4nJkH9Vj2Q_UfWpRe1mtRzd8s_Hrf8dCJYXsq9YdjjufQtNMCRbA'*/}}
  NAMESERVER_USER: {{$nameserverUsername}}
  NAMESERVER_PASSWORD: {{$nameserverPassword}}
  NAMESERVER_ENDPOINT: {{$nameserverEndpoint}}

  POD_CIDR: {{$podCidr}}
  SVC_CIDR: {{$svcCidr}}

  CLUSTER_ID: {{$clusterId}}
