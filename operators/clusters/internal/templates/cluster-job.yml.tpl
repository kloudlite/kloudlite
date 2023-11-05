{{- $jobName := get . "job-name" }} 
{{- $jobNamespace := get . "job-namespace" }} 
{{- $labels := get . "labels" | default dict }} 
{{- $ownerRefs := get . "owner-refs" |default list }}

{{- $serviceAccountName := get . "service-account-name" }} 

{{- $kubeconfigSecretName := get . "kubeconfig-secret-name" }}
{{- $kubeconfigSecretNamespace := get . "kubeconfig-secret-namespace" }}
{{- $kubeconfigSecreAnnotations := get . "kubeconfig-secret-annotations" }}

{{- $awsS3BucketName := get . "aws-s3-bucket-name" }} 
{{- $awsS3BucketFilepath := get . "aws-s3-bucket-filepath" }}
{{- $awsS3BucketRegion := get . "aws-s3-bucket-region" }} 

{{- $awsAccessKeyId := get . "aws-access-key-id" }}
{{- $awsSecretAccessKey := get . "aws-secret-access-key" }}

{{- $action := get . "action" }} 
{{- if not (or (eq $action "apply") (eq $action "delete")) }}
{{- fail "action should be either apply,delete" -}}
{{- end }}

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
    metadata:
      annotations:
        kloudlite.io/job_name: {{$jobName | squote}}
        kloudlite.io/job_type: "cluster-job"
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
          - name: HELM_CACHE_HOME
            value: ".helm-cache"
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
            
            if [ "{{$action}}" = "delete" ]; then
              terraform destroy --var-file ./values.json -auto-approve -no-color 2>&1 | tee /dev/termination-log
              kubectl delete secret -n {{$kubeconfigSecretNamespace}} {{$kubeconfigSecretName}} --ignore-not-found=true

            else
              terraform plan -out tfplan --var-file ./values.json -no-color 2>&1 | tee /dev/termination-log
              terraform apply -no-color tfplan 2>&1 | tee /dev/termination-log
              terraform state pull | jq '.outputs.kubeconfig.value' -r > kubeconfig

              kubectl apply -f - <<EOF
              apiVersion: v1
              kind: Secret
              metadata:
                name: {{$kubeconfigSecretName}}
                namespace: {{$kubeconfigSecretNamespace}}
                annotations: {{$kubeconfigSecreAnnotations | toYAML | nindent 18}}
              data:
                kubeconfig: $(cat kubeconfig)
                k3s_agent_token: $(terraform output -json k3s_agent_token | jq -r)
            EOF
            fi
            exit 0

      restartPolicy: Never
  backoffLimit: 1
