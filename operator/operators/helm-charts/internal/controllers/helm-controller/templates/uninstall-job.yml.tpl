{{- $name := get . "job-name" }} 
{{- $namespace := get . "job-namespace" }} 
{{- $labels := get . "labels" | default dict }} 
{{- $ownerRefs := get . "owner-refs" | default list }} 

{{- $serviceAccountName := get . "service-account-name" }} 
{{- $tolerations := get . "tolerations"  | default list }} 
{{- $affinity := get . "affinity" | default dict }}
{{- $nodeSelector := get . "node-selector" }} 
{{- $backoffLimit := get . "backoff-limit" | default 1 }} 

{{- $releaseName := get . "release-name" }} 
{{- $releaseNamespace := get . "release-namespace" }} 

{{- $preUninstall := get . "pre-uninstall" }}
{{- $postUninstall := get . "post-uninstall" }}

apiVersion: batch/v1
kind: Job
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  labels: {{$labels | toYAML | nindent 4}}
  ownerReferences: {{$ownerRefs | toYAML| nindent 4}}
spec:
  template:
    metadata:
      annotations:
        kloudlite.io/job_name: {{$name}}
        kloudlite.io/job_type: "helm-uninstall"
    spec:
      serviceAccountName: {{$serviceAccountName}}
      {{ if $tolerations }}
      tolerations: {{$tolerations | toYAML | nindent 10 }}
      {{ end }}
      {{- if $affinity }}
      affinity: {{$affinity | toYAML | nindent 10 }}
      {{- end }}
      {{- if $nodeSelector }}
      nodeSelector: {{$nodeSelector | toYAML | nindent 10}}
      {{- end }}
      containers:
      - name: helm
        {{- /* image: alpine/helm:3.12.3 */}}
        image: ghcr.io/kloudlite/job-runners/helm:v1.0.5-nightly
        command:
          - bash
          - -c
          - |+
            set -o pipefail

            {{- if $preUninstall }}
            echo "running pre-uninstall job script"
            {{ $preUninstall | nindent 12 }}
            {{- end }}

            helm uninstall --wait {{$releaseName}} --namespace {{$releaseNamespace}} 2>&1 | tee "/dev/termination-log"

            while true; do
              helm get values {{$releaseName}} -n {{$releaseNamespace}} > /dev/null 2>&1
              if [ $? -ne 0 ]; then
                echo "helm release successfully uninstalled"
                break
              fi
              echo "waiting for helm release to be uninstalled ..."
              sleep 1
            done

            {{- if $postUninstall }}
            echo "running post-uninstall job script"
            {{ $postUninstall | nindent 12 }}
            {{- end }}
      restartPolicy: Never
  backoffLimit: {{$backoffLimit | int}}
