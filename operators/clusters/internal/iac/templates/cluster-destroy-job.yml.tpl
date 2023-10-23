{{- $jobName := get . "job-name" }} 
{{- $jobNamespace := get . "job-namespace" }} 
{{- $labels := get . "labels" | default dict }} 
{{- $ownerRefs := get . "owner-refs" |default list }}

{{- $serviceAccountName := get . "service-account-name" }} 

{{- $kubeconfigSecretName := get . "kubeconfig-secret-name" }}
{{- $kubeconfigSecretNamespace := get . "kubeconfig-secret-namespace" }}

{{- $awsS3BucketName := get . "aws-s3-bucket-name" }} 
{{- $awsS3BucketFilepath := get . "aws-s3-bucket-filepath" }}
{{- $awsS3BucketRegion := get . "aws-s3-bucket-region" }} 

{{- $awsAccessKeyId := get . "aws-access-key-id" }}
{{- $awsSecretAccessKey := get . "aws-secret-access-key" }}

{{- $valuesJson := get . "values.json" }} 

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
      serviceAccountName: {{$serviceAccountName}}
      containers:
      - name: iac
        image: ghcr.io/kloudlite/infrastructure-as-code:v1.0.5-nightly-dev
        imagePullPolicy: Always
        env:
          - name: AWS_S3_BUCKET_NAME
            value: {{$awsS3BucketName}}
          - name: AWS_S3_BUCKET_FILEPATH
            value: {{$awsS3BucketFilepath}}
          - name: AWS_S3_BUCKET_REGION
            value: {{$awsS3BucketRegion}}
          - name: AWS_ACCESS_KEY_ID
            value: {{$awsAccessKeyId}}
          - name: AWS_SECRET_ACCESS_KEY
            value: {{$awsSecretAccessKey}}
        command:
          - bash
          - -c
          - |+
            set -o pipefail
            set -o errexit

            unzip $TERRAFORM_ZIPFILE

            pushd "$TEMPLATES_DIR/kl-target-cluster-aws-only-masters"

            envsubst < state-backend.tf.tpl > state-backend.tf
            
            cat > values.json <<EOF
            {{$valuesJson}}
            EOF

            terraform init -no-color 2>&1 | tee /dev/termination-log
            {{- /* terraform destroy -var-file ./values.json -var aws_nvidia_gpu_ami=asdfasdfdsaf -var enable_nvidia_gpu_support=false --auto-approve | tee /dev/termination-log */}}
            terraform destroy -var-file ./values.json --auto-approve | tee /dev/termination-log
            kubectl delete secret -n {{$kubeconfigSecretNamespace}} {{$kubeconfigSecretName}} --ignore-not-found=true
      restartPolicy: Never
  backoffLimit: 1
