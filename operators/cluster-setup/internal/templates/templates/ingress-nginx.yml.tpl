{{- $ingressValues := get . "ingress-values" -}}

{{- with $ingressValues }}
{{/*gotype: operators.kloudlite.io/apis/cluster-setup/v1.IngressValues*/}}
controller:
  ingressClassByName: true
  ingressClass: {{.ClassName}}
  electionID: {{.ClassName}}
  ingressClassResource:
    enabled: true
    name: "{{.ClassName}}"
    controllerValue: "k8s.io/{{.ClassName}}"

  hostNetwork: true
  hostPort:
    enabled: true
    ports:
      http: 80
      https: 443

  kind: DaemonSet

  service:
    type: "ClusterIP"

  {{- if .PodLabels }}
  podLabels: {{.PodLabels | toYAML | nindent 4}}
  {{- end}}

  resources:
    requests:
      cpu: {{.Resources.Cpu.Min}}
      memory: {{.Resources.Memory}}
    limits:
      cpu: {{.Resources.Cpu.Max}}
      memory: {{.Resources.Memory}}

  {{- if .NodeSelector }}
  nodeSelector: {{.NodeSelector | toYAML | nindent 4 }}
  {{- end -}}

  {{- if .Tolerations }}
  tolerations: {{.Tolerations | toYAML | nindent 4}}
  {{- end}}

  admissionWebhooks:
    failurePolicy: Ignore
{{- end }}
