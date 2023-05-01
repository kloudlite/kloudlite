apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.operators.wgOperator.name}}-env
  namespace: {{.Release.Namespace}}
stringData:
  NAMESERVER_ENDPOINT: {{.Values.wg.nameserverEndpoint}}
  NAMESERVER_USER: {{.Values.wg.nameserverUser}}
  NAMESERVER_PASSWORD: {{.Values.wg.nameserverPassword}}
  WG_DOMAIN: {{.Values.wg.wgDomain}}
  POD_CIDR: {{.Values.wg.podCidr}}
  SVC_CIDR: {{.Values.wg.svcCidr}}
