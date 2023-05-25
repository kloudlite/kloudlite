apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.operators.wgOperator.name}}-env
  namespace: {{.Release.Namespace}}
stringData:
  NAMESERVER_ENDPOINT: {{.Values.wg.nameserver.endpoint}}
  NAMESERVER_BASIC_AUTB_ENABLED: {{.Values.wg.nameserver.basicAuth.enabled |squote}}
  {{- if .Values.wg.nameserver.basicAuth.enabled }}
  NAMESERVER_USER: {{.Values.wg.nameserverUser}}
  NAMESERVER_PASSWORD: {{.Values.wg.nameserverPassword}}
  {{- end }}
  WG_DOMAIN: {{.Values.wg.baseDomain}}
  POD_CIDR: {{.Values.wg.podCidr}}
  SVC_CIDR: {{.Values.wg.svcCidr}}
