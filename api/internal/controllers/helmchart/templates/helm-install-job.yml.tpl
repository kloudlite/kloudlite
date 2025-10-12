{{- with .}}
{{- /* gotype: github.com/kloudlite/plugin-helm-chart/internal/controller/templates.InstallJobVars */ -}}
backoffLimit: {{ .BackOffLimit | default 1 | int}}
template:
  metadata:
    labels: {{.PodLabels | default dict | toJson }}
    annotations: {{.PodAnnotations | default dict | toJson }}
  spec:
    serviceAccountName: {{.ServiceAccountName | toJson }}
    tolerations: {{ .Tolerations | default list | toJson }}
    affinity: {{ .Affinity | default dict | toJson }}
    nodeSelector: {{ .NodeSelector | default dict | toJson }}
    containers:
      - name: helm
        image: {{.Image}}
        imagePullPolicy: {{ .ImagePullPolicy | default (hasSuffix .Image "-nightly" | ternary "Always" "IfNotPresent") }}
        command:
          - bash
          - "-c"
          - |+ #bash
            set -o nounset
            set -o pipefail
            set -o errexit

            {{- if .PreInstall }}
            echo "[#] running pre-install script"
            {{ .PreInstall | nindent 12 }}
            {{- end }}

            echo "[#] adding helm repository"
            helm repo add helm-repo {{.ChartRepoURL}}

            echo "[#] updating helm repository"
            helm repo update helm-repo

            cat > values.yml <<EOF
            {{ .HelmValuesYAML | nindent 12 }}
            EOF

            version_args=()
            if [ -n "{{.ChartVersion}}" ]; then
              version_args+=("--version" "{{.ChartVersion}}")
            fi

            echo "[#] installing/updating helm resource"
            helm upgrade --wait --install {{.ReleaseName}} helm-repo/{{.ChartName}} --namespace {{.ReleaseNamespace}} ${version_args[@]} --values values.yml 2>&1 | tee /dev/termination-log

            {{- if .PostInstall }}
            echo "[#] running post-install script"
            {{ .PostInstall | nindent 12 }}
            {{- end }}
    restartPolicy: Never
{{- end }}
