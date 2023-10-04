{{- $jobName := get . "job-name" }} 
{{- $jobNamespace := get . "job-namespace" }} 
{{- $labels := get . "labels" | default dict }} 
{{- $ownerRefs := get . "owner-refs" |default list }}

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
      containers:
      - name: iac
        image: ghcr.io/kloudlite/infrastructure-as-code:v1.0.5-nightly
        imagePullPolicy: Always
        env:
          - name: AWS_S3_BUCKET_NAME
            value: {{$awsS3BucketName}}
          - name: AWS_S3_BUCKET_FILEPATH
            value: {{$awsS3BucketFilepath}}
          - name: AWS_S3_BUCKET_REGION
            value: {{$awsS3BucketRegion}}
          {{- if $awsAccessKeyId }}
          - name: AWS_ACCESS_KEY_ID
            value: {{$awsAccessKeyId}}
          {{- end }}
          {{- if $awsSecretAccessKey }}
          - name: AWS_SECRET_ACCESS_KEY
            value: {{$awsSecretAccessKey}}
          {{- end }}
        command:
          - bash
          - -c
          - |+
            set -o pipefail
            set -o errexit

            unzip terraform.zip && rm terraform.zip

            pushd "infrastructures/templates/aws-s3-bucket"

            envsubst < terraform.tf.tpl > terraform.tf
            
            cat > values.json <<EOF
            {{$valuesJson}}
            EOF

            terraform init -no-color 2>&1 | tee /dev/termination-log
            terraform plan --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
            terraform apply -no-color tfplan 2>&1 | tee /dev/termination-log

            
      restartPolicy: Never
  backoffLimit: 1
