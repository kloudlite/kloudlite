{{- $jobName := get . "job-name" }} 
{{- $jobNamespace := get . "job-namespace" }} 
{{- $labels := get . "labels" | default dict }} 
{{- $ownerRefs := get . "owner-refs" |default list }}

{{- $repoUrl := get . "repo-url" }}
{{- $repoName := get . "repo-name" }} 

{{- $chartName := get . "chart-name" }} 
{{- $chartVersion := get . "chart-version" }} 

{{- $releaseName := get . "release-name" }} 
{{- $releaseNamespace := get . "release-namespace" }} 
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
    spec:
      serviceAccountName: kloudlite-cluster-svc-account
      containers:
      - name: helm
        image: alpine/helm:3.12.3
        command:
          - bash
          - -c
          - |+
            set -o pipefail

            helm repo add {{$repoName}} {{$repoUrl}}
            helm repo update {{$repoName}}
            cat > values.yml <<EOF
            {{ $valuesYaml | nindent 12 }}
            EOF
            helm upgrade --install {{$releaseName}} {{$chartName}} --namespace {{$releaseNamespace}} --version {{$chartVersion}} --values values.yml 2>&1 | tee /dev/termination-log
      restartPolicy: Never
  backoffLimit: 4
