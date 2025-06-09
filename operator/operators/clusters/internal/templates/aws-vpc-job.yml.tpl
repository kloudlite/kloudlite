{{- /*gotype: github.com/kloudlite/operator/operators/clusters/internal/templates.AwsVPCJobVars*/ -}}
{{ with .}}
apiVersion: crds.kloudlite.io/v1
kind: Lifecycle
metadata: {{.JobMetadata | toYAML | nindent 2}}
spec:
  onApply:
    backOffLimit: 0
    podSpec:
      restartPolicy: Never
      containers:
        - name: iac
          image: {{.JobImage}}
          imagePullPolicy: "Always"

          resources:
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
              value: {{.TFWorkspaceSecretNamespace | squote}}

            {{- /* # since we are not using Assume Role Model, we do not need it */}}
            {{- /* - name: AWS_ACCESS_KEY_ID */}}
            {{- /*   value: {{.AWS.AccessKeyID }} */}}
            {{- /**/}}
            {{- /* - name: AWS_SECRET_ACCESS_KEY */}}
            {{- /*   value: {{.AWS.AccessKeySecret}} */}}

          command:
            - bash
            - -c
            - |+
              set -o pipefail
              set -o errexit

              eval $DECOMPRESS_CMD

              pushd "$TEMPLATES_DIR/aws/vpc"

              envsubst < state-backend.tf.tpl > state-backend.tf

              terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
              terraform workspace select --or-create {{.TFWorkspaceName}}

              cat > values.json <<EOF
              {{.ValuesJSON}}
              EOF

              terraform plan -parallelism=2 -out tfplan --var-file ./values.json -no-color 2>&1 | tee /dev/termination-log
              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log

              terraform state pull | jq '.outputs' -r > outputs.json

              kubectl apply -f - <<EOF
              apiVersion: v1
              kind: Secret
              metadata:
                name: {{.VPCOutputSecretName}}
                namespace: {{.VPCOutputSecretNamespace}}
              data:
                vpc_id: $(cat outputs.json | jq -r '.vpc_id.value' | base64 | tr -d '\n')
                vpc_public_subnets: $(cat outputs.json | jq -r '.vpc_public_subnets.value' |base64| tr -d '\n')
              EOF
              exit 0

  onDelete:
    podSpec:
      restartPolicy: Never
      containers:
        - name: iac
          image: {{.JobImage}}
          imagePullPolicy: "Always"

          resources:
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
              value: {{.TFWorkspaceSecretNamespace | squote}}

            - name: AWS_ACCESS_KEY_ID
              value: {{ .AWS.AccessKeyID }}

            - name: AWS_SECRET_ACCESS_KEY
              value: {{.AWS.AccessKeySecret}}

          command:
            - bash
            - -c
            - |+
              set -o pipefail
              set -o errexit

              eval $DECOMPRESS_CMD

              pushd "$TEMPLATES_DIR/aws/vpc"

              envsubst < state-backend.tf.tpl > state-backend.tf

              terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
              terraform workspace select --or-create {{.TFWorkspaceName}}

              cat > values.json <<EOF
              {{.ValuesJSON}}
              EOF

              terraform plan -parallelism=2 --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log

              kubectl delete secret/{{.VPCOutputSecretName}} -n {{.VPCOutputSecretNamespace}} --ignore-not-found
              exit 0
{{end}}

{{/*apiVersion: batch/v1*/}}
{{/*kind: Job*/}}
{{/*metadata:*/}}
{{/*  name: {{$jobName}}*/}}
{{/*  namespace: {{$jobNamespace}}*/}}
{{/*  labels: {{$labels | toYAML | nindent 4}}*/}}
{{/*  ownerReferences: {{$ownerRefs | toYAML| nindent 4}}*/}}
{{/*spec:*/}}
{{/*  template:*/}}
{{/*    metadata:*/}}
{{/*      annotations: {{$podAnnotations | toYAML | nindent 8 }}*/}}
{{/*    spec:*/}}
{{/*      serviceAccountName: {{$serviceAccountName}}*/}}
{{/*      nodeSelector: {{$jobNodeSelector | toYAML | nindent 8 }}*/}}
{{/*      tolerations: {{$jobTolerations | toYAML | nindent 8 }}*/}}
{{/*      containers:*/}}
{{/*      - name: iac*/}}
{{/*        image: {{$jobImage}}*/}}
{{/*        imagePullPolicy: Always*/}}

{{/*        resources:*/}}
{{/*          {{- /* requests: */}}
{{/*          {{- /*   cpu: 500m */}}
{{/*          {{- /*   memory: 1000Mi */}}
{{/*          {{- /* limits: */}}
{{/*          {{- /*   cpu: 500m */}}
{{/*          {{- /*   memory: 1000Mi */}}
{{/*          requests:*/}}
{{/*            cpu: 400m*/}}
{{/*            memory: 400Mi*/}}
{{/*          limits:*/}}
{{/*            cpu: 400m*/}}
{{/*            memory: 400Mi*/}}

{{/*        env:*/}}
{{/*          - name: KUBE_IN_CLUSTER_CONFIG*/}}
{{/*            value: "true"*/}}

{{/*          - name: KUBE_NAMESPACE*/}}
{{/*            value: {{$tfStateSecretNamespace | squote}}*/}}

{{/*          - name: AWS_ACCESS_KEY_ID*/}}
{{/*            value: {{$awsAccessKeyId}}*/}}

{{/*          - name: AWS_SECRET_ACCESS_KEY*/}}
{{/*            value: {{$awsSecretAccessKey}}*/}}

{{/*        command:*/}}
{{/*          - bash*/}}
{{/*          - -c*/}}
{{/*          - |+*/}}
{{/*            set -o pipefail*/}}
{{/*            set -o errexit*/}}

{{/*            eval $DECOMPRESS_CMD*/}}

{{/*            pushd "$TEMPLATES_DIR/aws-vpc"*/}}

{{/*            envsubst < state-backend.tf.tpl > state-backend.tf*/}}

{{/*            terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log*/}}
{{/*            terraform workspace select --or-create {{$tfStateSecretName}} */}}
{{/*            */}}
{{/*            cat > values.json <<EOF*/}}
{{/*            {{$valuesJson}}*/}}
{{/*            EOF*/}}

{{/*            if [ "{{$action}}" = "delete" ]; then*/}}
{{/*              {{- /* terraform destroy --var-file ./values.json -auto-approve -no-color 2>&1 | tee /dev/termination-log */}}
{{/*              terraform plan -parallelism=2 --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log*/}}
{{/*              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log*/}}

{{/*              kubectl delete secret/{{$vpcOutputSecretName}} -n {{$vpcOutputSecretNamespace}} --ignore-not-found*/}}
{{/*            else*/}}
{{/*              terraform plan -parallelism=2 -out tfplan --var-file ./values.json -no-color 2>&1 | tee /dev/termination-log*/}}
{{/*              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log*/}}

{{/*              terraform state pull | jq '.outputs' -r > outputs.json*/}}

{{/*              kubectl apply -f - <<EOF*/}}
{{/*              apiVersion: v1*/}}
{{/*              kind: Secret*/}}
{{/*              metadata:*/}}
{{/*                name: {{$vpcOutputSecretName}}*/}}
{{/*                namespace: {{$vpcOutputSecretNamespace}}*/}}
{{/*              data:*/}}
{{/*                vpc_id: $(cat outputs.json | jq -r '.vpc_id.value' | base64 | tr -d '\n')*/}}
{{/*                vpc_public_subnets: $(cat outputs.json | jq -r '.vpc_public_subnets.value' |base64| tr -d '\n')*/}}
{{/*            EOF*/}}
{{/*            fi*/}}
{{/*            exit 0*/}}

{{/*      restartPolicy: Never*/}}
{{/*  backoffLimit: 1*/}}
