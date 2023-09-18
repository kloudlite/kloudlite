{{- $name := get . "job-name" }} 
{{- $namespace := get . "job-namespace" }} 
{{- $labels := get . "labels" | default dict }} 
{{- $ownerRefs := get . "owner-refs" | default list }} 

{{- $releaseName := get . "release-name" }} 
{{- $releaseNamespace := get . "release-namespace" }} 

apiVersion: batch/v1
kind: Job
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
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
            helm uninstall --wait {{$releaseName}} --namespace {{$releaseNamespace}} 2>&1 | tee "/dev/termination-log"

            while true; do
              helm get values {{$releaseName}} -n {{$releaseNamespace}}
              if [ $? -ne 0 ]; then
                echo "helm release successfully uninstalled"
                break
              fi
              echo "waiting for helm release to be uninstalled ..."
              sleep 1
            done
      restartPolicy: Never
  backoffLimit: 4
