{{- $namespace := get . "namespace" -}}
{{- $podCidr := get . "pod-cidr" | default "10.42.0.0/16" -}}
{{- $svcCidr := get . "svc-cidr" | default "10.43.0.0/16" -}}
{{- $wildcardDomain := get . "wildcard-domain" -}}
{{- $ownerRefs := get . "owner-refs" | default list -}}

{{- $nameserverEndpoint := get . "nameserver-endpoint" -}}
{{- $nameserverUsername := get . "nameserver-username" -}}
{{- $nameserverPassword := get . "nameserver-password" -}}

apiVersion: v1
kind: Secret
metadata:
  name: wg-operator-env
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toYAML | nindent 4}}
stringData:
  NAMESERVER_USER: {{$nameserverUsername}}
  NAMESERVER_PASSWORD: {{$nameserverPassword}}
  NAMESERVER_ENDPOINT: {{$nameserverEndpoint}}

  WG_DOMAIN: '{{$wildcardDomain}}'
  POD_CIDR: {{$podCidr}}
  SVC_CIDR: {{$svcCidr}}
