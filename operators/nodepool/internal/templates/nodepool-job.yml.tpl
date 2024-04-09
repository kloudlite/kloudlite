{{- $jobName := get . "job-name" }} 
{{- $jobNamespace := get . "job-namespace" }} 

{{- $iacJobImage := get . "iac-job-image" }}

{{- $labels := get . "labels" | default dict }} 
{{- $annotations := get . "annotations" | default dict }}

{{- $podAnnotations := get . "pod-annotations" | default dict }}

{{- $ownerRefs := get . "owner-refs" | default list }}

{{- $jobNodeSelector := get . "job-node-selector" }} 

{{- $nodepoolName := get . "nodepool-name" }}
{{- $tfStateSecretNamespace := get . "tfstate-secret-namespace" }}

{{- $action := get . "action" }} 
{{- if (not (or (eq $action "apply") (eq $action "delete"))) }}
{{- fail "action must be either apply or delete" }}
{{- end }}

{{- $valuesJson := get . "values.json" }} 

{{- $cloudprovider := get . "cloudprovider" }}

apiVersion: batch/v1
kind: Job
metadata:
  name: {{$jobName}}
  namespace: {{$jobNamespace}}
  labels: {{$labels | toYAML | nindent 4}}
  annotations: {{$annotations | toYAML | nindent 4 }}
  ownerReferences: {{$ownerRefs | toYAML| nindent 4}}
spec:
  template:
    metadata:
      annotations: {{$podAnnotations | toYAML | nindent 8 }}
      labels: {{$labels | toYAML | nindent 8 }}
    spec:
      tolerations:
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
          operator: Exists

      nodeSelector: {{$jobNodeSelector | toYAML | nindent 10}}
      serviceAccountName: "kloudlite-jobs"

      resources:
        requests:
          cpu: 500m
          memory: 1000Mi
        limits:
          cpu: 500m
          memory: 1000Mi

      containers:
      - name: main
        image: {{$iacJobImage}}
        imagePullPolicy: Always
        env:
          - name: KUBE_IN_CLUSTER_CONFIG
            value: "true"

          - name: KUBE_NAMESPACE
            value: {{$tfStateSecretNamespace | squote}}
        command:
          - bash
          - -c
          - |+
            set -o pipefail
            set -o errexit

            eval $DECOMPRESS_CMD

            tdir=""
            if [ "{{$cloudprovider}}" = "aws" ]; then
              tdir="$TEMPLATES_DIR/kl-target-cluster-aws-only-workers"
            fi
            if [ "{{$cloudprovider}}" = "gcp" ]; then
              tdir="$TEMPLATES_DIR/gcp/worker-nodes"
            fi

            pushd "$tdir"

            envsubst < state-backend.tf.tpl > state-backend.tf

            terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
            terraform workspace select --or-create {{$nodepoolName}} 
            
            cat > values.json <<EOF
            {{$valuesJson}}
            EOF
            
            terraform init -no-color 2>&1 | tee /dev/termination-log
            if [ "{{$action}}" = "apply" ]; then
              terraform plan -parallelism=2 --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
            else
              terraform plan -parallelism=2 --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
            fi
            
            terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log
      restartPolicy: Never
  backoffLimit: 1
