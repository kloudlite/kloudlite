{{- $jobName := get . "job-name" }}
{{- $jobNamespace := get . "job-namespace" }}

{{- $labels := get . "labels" | default dict }}

{{- $jobTolerations := get . "job-tolerations" | default list }}
{{- $jobNodeSelector := get . "job-node-selector"  | default dict }}

{{- $podAnnotations := get . "pod-annotations" | default dict }}

{{- $ownerRefs := get . "owner-refs" |default list }}

{{- $serviceAccountName := get . "service-account-name" }} 

{{- $kubeconfigSecretName := get . "kubeconfig-secret-name" }}
{{- $kubeconfigSecretNamespace := get . "kubeconfig-secret-namespace" }}
{{- $kubeconfigSecreAnnotations := get . "kubeconfig-secret-annotations" }}

{{- $clusterName := get . "cluster-name" }}
{{- $tfStateSecretNamespace := get . "tf-state-secret-namespace" }}

{{- $awsAccessKeyId := get . "aws-access-key-id" }}
{{- $awsSecretAccessKey := get . "aws-secret-access-key" }}

{{- $action := get . "action" }}
{{- if not (or (eq $action "apply") (eq $action "delete")) }}
{{- fail "action should be either apply,delete" -}}
{{- end }}

{{- $valuesJson := get . "values.json" }} 

{{- $jobImage := get . "job-image" }}

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
          requests:
            cpu: 500m
            memory: 1000Mi
          limits:
            cpu: 500m
            memory: 1000Mi

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

            pushd "$TEMPLATES_DIR/kl-target-cluster-aws-only-masters"
            envsubst < state-backend.tf.tpl > state-backend.tf

            terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
            terraform workspace select --or-create {{$clusterName}} 
            
            cat > values.json <<EOF
            {{$valuesJson}}
            EOF

            if [ "{{$action}}" = "delete" ]; then
              {{- /* terraform destroy --var-file ./values.json -auto-approve -no-color 2>&1 | tee /dev/termination-log */}}
              terraform plan --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log
              kubectl delete secret/{{$kubeconfigSecretName}} -n {{$kubeconfigSecretNamespace}} --ignore-not-found=true
            else
              terraform plan -out tfplan --var-file ./values.json -no-color 2>&1 | tee /dev/termination-log
              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log

              terraform state pull | jq '.outputs' -r > outputs.json
              
              cat outputs.json

              kubectl apply -f - <<EOF
              apiVersion: v1
              kind: Secret
              metadata:
                name: {{$kubeconfigSecretName}}
                namespace: {{$kubeconfigSecretNamespace}}
                annotations: {{$kubeconfigSecreAnnotations | toYAML | nindent 18}}
              data:
                kubeconfig: $(cat outputs.json | jq '.kubeconfig.value')
                k3s_params: $(cat outputs.json | jq -r '."kloudlite-k3s-params".value' | base64 | tr -d '\n')
                k3s_agent_token: $(cat outputs.json | jq -r '.k3s_agent_token.value' | base64 | tr -d '\n')
            EOF
            fi
            exit 0

      restartPolicy: Never
  backoffLimit: 1
