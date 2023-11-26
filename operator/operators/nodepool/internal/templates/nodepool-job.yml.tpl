{{- $jobName := get . "job-name" }} 
{{- $jobNamespace := get . "job-namespace" }} 

{{- $labels := get . "labels" | default dict }} 
{{- $annotations := get . "annotations" | default dict }} 
{{- $ownerRefs := get . "owner-refs" | default list }}

{{- $jobNodeSelector := get . "job-node-selector" }} 

{{- /* {{- $serviceAccountName := get . "service-account-name" }}  */}}

{{- $awsS3BucketName := get . "aws-s3-bucket-name" }} 
{{- $awsS3BucketFilepath := get . "aws-s3-bucket-filepath" }}
{{- $awsS3BucketRegion := get . "aws-s3-bucket-region" }} 

{{- $awsS3AccessKey := get . "aws-s3-access-key" }} 
{{- $awsS3SecretKey := get . "aws-s3-secret-key" }} 

{{- $action := get . "action" }} 
{{- if (not (or (eq $action "apply") (eq $action "delete"))) }}
{{- fail "action must be either apply or delete" }}
{{- end }}

{{- $valuesJson := get . "values.json" }} 

apiVersion: batch/v1
kind: Job
metadata:
  name: {{$jobName}}
  namespace: {{$jobNamespace}}
  labels: {{$labels | toYAML | nindent 4}}
  annotations: {{$annotations | toYAML | nindent 4}}
  ownerReferences: {{$ownerRefs | toYAML| nindent 4}}
spec:
  template:
    metadata:
      annotations:
        kloudlite.io/job_name: {{$jobName}}
        kloudlite.io/job_type: nodepool-{{$action}}
    spec:
      tolerations:
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
          operator: Exists

      nodeSelector: {{$jobNodeSelector | toYAML | nindent 10}}

      containers:
      - name: main
        image: ghcr.io/kloudlite/infrastructure-as-code:v1.0.5-nightly-dev
        imagePullPolicy: Always
        env:
          - name: AWS_S3_BUCKET_NAME
            value: {{$awsS3BucketName}}
          - name: AWS_S3_BUCKET_FILEPATH
            value: {{$awsS3BucketFilepath}}
          - name: AWS_S3_BUCKET_REGION
            value: {{$awsS3BucketRegion}}
          {{- if $awsS3AccessKey }}
          - name: AWS_ACCESS_KEY
            value: {{$awsS3AccessKey}}
          {{- end }}
          {{- if $awsS3SecretKey }}
          - name: AWS_SECRET_KEY
            value: {{$awsS3SecretKey}}
          {{- end }}
        command:
          - bash
          - -c
          - |+
            set -o pipefail
            set -o errexit

            unzip $TERRAFORM_ZIPFILE

            pushd "$TEMPLATES_DIR/kl-target-cluster-aws-only-workers"

            envsubst < state-backend.tf.tpl > state-backend.tf
            
            cat > values.json <<EOF
            {{$valuesJson}}
            EOF
            
            terraform init -no-color 2>&1 | tee /dev/termination-log
            if [ "{{$action}}" = "apply" ]; then
              terraform plan --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
            else
              terraform plan --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
            fi
            terraform apply -no-color tfplan 2>&1 | tee /dev/termination-log
      restartPolicy: Never
  backoffLimit: 1
