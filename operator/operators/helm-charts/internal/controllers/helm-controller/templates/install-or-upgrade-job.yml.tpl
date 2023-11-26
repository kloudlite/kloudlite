{{- $jobName := get . "job-name" }} 
{{- $jobNamespace := get . "job-namespace" }} 
{{- $labels := get . "labels" | default dict }} 
{{- $ownerRefs := get . "owner-refs" |default list }}

{{- $serviceAccountName := get . "service-account-name" }} 
{{- $tolerations := get . "tolerations"  | default list }} 
{{- $affinity := get . "affinity" | default dict }}
{{- $nodeSelector := get . "node-selector" }} 
{{- $backoffLimit := get . "backoff-limit" | default 1 }} 

{{- $repoUrl := get . "repo-url" }}
{{- $repoName := get . "repo-name" }} 

{{- $chartName := get . "chart-name" }} 
{{- $chartVersion := get . "chart-version" }} 

{{- $releaseName := get . "release-name" }} 
{{- $releaseNamespace := get . "release-namespace" }} 

{{- $preInstall := get . "pre-install" }}
{{- $postInstall := get . "post-install" }}
{{- $valuesYaml := get . "values-yaml" }} 

apiVersion: batch/v1
kind: Job
metadata:
  name: {{$jobName}}
  namespace: {{$jobNamespace}}
  labels: {{$labels | toYAML | nindent 4}}
  ownerReferences: {{$ownerRefs | toYAML| nindent 4}}
spec:
  template:
    metadata:
      annotations:
        kloudlite.io/job_name: {{$jobName}}
        kloudlite.io/job_type: "helm-install"
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
        imagePullPolicy: Always
        command:
          - bash
          - -c
          - |+
            set -o pipefail

            helm repo add {{$repoName}} {{$repoUrl}}
            helm repo update {{$repoName}}

            {{- if $preInstall }}
            echo "running pre-install job script"
            {{ $preInstall | nindent 12 }}
            {{- end }}

            cat > values.yml <<EOF
            {{ $valuesYaml | nindent 12 }}
            EOF

            helm upgrade --install {{$releaseName}} {{$chartName}} --namespace {{$releaseNamespace}} --version {{$chartVersion}} --values values.yml 2>&1 | tee /dev/termination-log

            {{- if $postInstall }}
            echo "running post-install job script"
            {{ $postInstall | nindent 12 }}
            {{- end }}
            
      restartPolicy: Never
  backoffLimit: {{$backoffLimit | int}}
