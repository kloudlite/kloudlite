{{- with .}}
{{- /* gotype: github.com/kloudlite/plugin-helm-chart/internal/controller/templates.UnInstallJobVars */ -}}
backoffLimit: {{ .BackOffLimit | default 1 | int}}
template:
  metadata:
    labels: {{.PodLabels | toJson}}
    annotations: {{.PodAnnotations | toJson }}
  spec:
    serviceAccountName: {{.ServiceAccountName | toJson }}
    tolerations: {{.Tolerations | toJson }}
    affinity: {{.Affinity | toJson}}
    nodeSelector: {{ .NodeSelector | toJson }}
    containers:
      - name: helm
        image: {{.Image}}
        imagePullPolicy: {{ .ImagePullPolicy | default (hasSuffix .Image "-nightly" | ternary "Always" "IfNotPresent" ) }}
        command:
          - bash
          - -c
          - |+ #bash
            set -o nounset
            set -o pipefail
            set -o errexit

            {{- if .PreUninstall }}
            echo "running pre-uninstall job script"
            {{ .PreUninstall | nindent 12 }}
            {{- end }}

            helm repo add helm-repo {{.ChartRepoURL}}
            helm repo update helm-repo

            helm uninstall {{.ReleaseName}} --namespace {{.ReleaseNamespace}} | tee /dev/termination-log

            {{- if .PostUninstall }}
            echo "running post-uninstall job script"
            {{ .PostUninstall | nindent 12 }}
            {{- end }}
          
    restartPolicy: Never
{{- end }}
