{{- /*gotype: github.com/kloudlite/operator/operators/clusters/internal/templates.GcpVPCJobVars*/ -}}
{{ with . }}
apiVersion: crds.kloudlite.io/v1
kind: Lifecycle
metadata:
  name: {{.JobMetadata.Name}}
  namespace: {{.JobMetadata.Namespace}}
  labels: {{.JobMetadata.Labels | toYAML | nindent 4}}
  annotations: {{.JobMetadata.Annotations | toYAML | nindent 4}}
  ownerReferences: {{.JobMetadata.OwnerReferences | toYAML | nindent 4}}
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
              value: {{.TFStateSecretNamespace | squote}}

          command:
            - bash
            - -c
            - |+
              set -o pipefail
              set -o errexit

              eval $DECOMPRESS_CMD

              {{- /* pushd "$TEMPLATES_DIR/{{.CloudProvider}}/vpc" */}}
              pushd "$TEMPLATES_DIR/gcp/vpc"

              envsubst < state-backend.tf.tpl > state-backend.tf

              terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
              terraform workspace select --or-create {{.TFStateSecretName}}

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
                vpc_name: $(cat outputs.json | jq -r '.vpc_name.value' | base64 | tr -d '\n')
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
              value: {{.TFStateSecretNamespace | squote}}

          command:
            - bash
            - -c
            - |+
              set -o pipefail
              set -o errexit

              eval $DECOMPRESS_CMD

              {{- /* pushd "$TEMPLATES_DIR/{{.CloudProvider}}/vpc" */}}
              pushd "$TEMPLATES_DIR/gcp/vpc"

              envsubst < state-backend.tf.tpl > state-backend.tf

              terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
              terraform workspace select --or-create {{.TFStateSecretName}}

              cat > values.json <<EOF
              {{.ValuesJSON}}
              EOF

              terraform plan -parallelism=2 --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log

              kubectl delete secret/{{.VPCOutputSecretName}} -n {{.VPCOutputSecretNamespace}} --ignore-not-found
              exit 0
      {{ end }}
