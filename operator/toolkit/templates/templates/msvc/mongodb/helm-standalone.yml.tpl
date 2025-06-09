{{- $obj :=  get . "object" }}
{{- $ownerRefs := get . "owner-refs"}}
{{- $storageClass := get . "storage-class"}}
{{- $freeze := get . "freeze" | default false}}
{{- $priorityClassName := get . "priority-class-name"  | default "stateful" -}}
{{- $existingSecret := get . "existing-secret" -}}

{{- $labels := get . "labels" | default dict }}

{{- with $obj }}
{{- /* gotype: github.com/kloudlite/operator/apis/mongodb.msvc/v1.StandaloneService*/ -}}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
	name: {{.Name}}
	namespace: {{.Namespace}}
	labels: {{ $labels | toYAML | nindent 4 }}
	ownerReferences: {{ $ownerRefs | toYAML | nindent 4 }}
spec:
	chartRepo:
		url: https://charts.bitnami.com/bitnami
		name: bitnami
	chartVersion: 13.18.1
	chartName: bitnami/mongodb

	valuesYaml: |+
		# source: https://github.com/bitnami/charts/tree/main/bitnami/mongodb/
		global:
			storageClass: {{$storageClass}}
		fullnameOverride: {{.Name}}
		image:
			tag: 5.0.8-debian-10-r20

		architecture: standalone
		useStatefulSet: true

		replicaCount: {{ if $freeze}}0{{else}}{{.Spec.ReplicaCount}}{{end}}

		commonLabels: {{$labels | toYAML | nindent 6}}
		podLabels: {{$labels | toYAML | nindent 6}}

		auth:
		  enabled: true
		  existingSecret: {{$existingSecret}}

		persistence:
		  enabled: true
		  size: {{.Spec.Resources.Storage.Size}}

		volumePermissions:
		  enabled: true

		metrics:
		  enabled: true

		livenessProbe:
		  enabled: true
		  timeoutSeconds: 15

		readinessProbe:
		  enabled: true
		  initialDelaySeconds: 10
		  periodSeconds: 30
		  timeoutSeconds: 20

		resources:
			requests:
				cpu: {{.Spec.Resources.Cpu.Min}}
				memory: {{.Spec.Resources.Memory}}
			limits:
				cpu: {{.Spec.Resources.Cpu.Max}}
				memory: {{.Spec.Resources.Memory}}
{{- end}}