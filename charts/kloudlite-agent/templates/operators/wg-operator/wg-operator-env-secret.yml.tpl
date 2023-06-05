apiVersion: v1
kind: Secret
metadata:
  name: {{.Values.operators.wgOperator.name}}-env
  namespace: {{.Release.Namespace}}
stringData:
  NAMESERVER_ENDPOINT: {{.Values.operators.wgOperator.configuration.nameserver.endpoint}}
  NAMESERVER_BASIC_AUTH_ENABLED: {{.Values.operators.wgOperator.configuration.nameserver.basicAuth.enabled |squote}}
  {{- if .Values.operators.wgOperator.configuration.nameserver.basicAuth.enabled }}
  NAMESERVER_USER: {{.Values.operators.wgOperator.configuration.nameserverUser}}
  NAMESERVER_PASSWORD: {{.Values.operators.wgOperator.configuration.nameserverPassword}}
  {{- end }}
  WG_DOMAIN: {{.Values.operators.wgOperator.configuration.baseDomain}}
  POD_CIDR: {{.Values.operators.wgOperator.configuration.podCidr}}
  SVC_CIDR: {{.Values.operators.wgOperator.configuration.svcCidr}}
