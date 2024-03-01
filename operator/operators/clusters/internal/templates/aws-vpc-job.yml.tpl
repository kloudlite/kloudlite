{{- $jobName := get . "job-name" }}
{{- $jobNamespace := get . "job-namespace" }}
{{- $jobImage := get . "job-image" }}
{{- $jobTolerations := get . "job-tolerations" | default list }}
{{- $jobNodeSelector := get . "job-node-selector"  | default dict }}

{{- $labels := get . "labels" | default dict }}

{{- $podAnnotations := get . "pod-annotations" | default dict }}

{{- $ownerRefs := get . "owner-refs" |default list }}

{{- $serviceAccountName := get . "service-account-name" }} 

{{- $tfStateSecretName := get . "tf-state-secret-name" }}
{{- $tfStateSecretNamespace := get . "tf-state-secret-namespace" }}

{{- $awsAccessKeyId := get . "aws-access-key-id" }}
{{- $awsSecretAccessKey := get . "aws-secret-access-key" }}

{{- $action := get . "action" }}
{{- if not (or (eq $action "apply") (eq $action "delete")) }}
{{- fail "action should be either apply,delete" -}}
{{- end }}

{{- $valuesJson := get . "values.json" }} 

{{- $vpcOutputSecretName := get . "vpc-output-secret-name" }}
{{- $vpcOutputSecretNamespace := get . "vpc-output-secret-namespace" }}

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
      annotations: {{$podAnnotations | toYAML | nindent 8 }}
    spec:
      serviceAccountName: {{$serviceAccountName}}
      nodeSelector: {{$jobNodeSelector | toYAML | nindent 8 }}
      tolerations: {{$jobTolerations | toYAML | nindent 8 }}
      containers:
      - name: iac
        image: {{$jobImage}}
        imagePullPolicy: Always

        resources:
          {{- /* requests: */}}
          {{- /*   cpu: 500m */}}
          {{- /*   memory: 1000Mi */}}
          {{- /* limits: */}}
          {{- /*   cpu: 500m */}}
          {{- /*   memory: 1000Mi */}}
          requests:
            cpu: 400m
            memory: 400Mi
          limits:
            cpu: 400m
            memory: 400Mi

        env:
          - name: KUBE_IN_CLUSTER_CONFIG
            value: "true"

          - name: KUBE_NAMESPACE
            value: {{$tfStateSecretNamespace | squote}}

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

            pushd "$TEMPLATES_DIR/aws-vpc"

            envsubst < state-backend.tf.tpl > state-backend.tf

            terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
            terraform workspace select --or-create {{$tfStateSecretName}} 
            
            cat > values.json <<EOF
            {{$valuesJson}}
            EOF

            if [ "{{$action}}" = "delete" ]; then
              {{- /* terraform destroy --var-file ./values.json -auto-approve -no-color 2>&1 | tee /dev/termination-log */}}
              terraform plan --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log

              kubectl delete secret/{{$vpcOutputSecretName}} -n {{$vpcOutputSecretNamespace}} --ignore-not-found
            else
              terraform plan -out tfplan --var-file ./values.json -no-color 2>&1 | tee /dev/termination-log
              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log

              terraform state pull | jq '.outputs' -r > outputs.json

              kubectl apply -f - <<EOF
              apiVersion: v1
              kind: Secret
              metadata:
                name: {{$vpcOutputSecretName}}
                namespace: {{$vpcOutputSecretNamespace}}
              data:
                vpc_id: $(cat outputs.json | jq -r '.vpc_id.value' | base64 | tr -d '\n')
                vpc_public_subnets: $(cat outputs.json | jq -r '.vpc_public_subnets.value' |base64| tr -d '\n')
            EOF
            fi
            exit 0

      restartPolicy: Never
  backoffLimit: 1
