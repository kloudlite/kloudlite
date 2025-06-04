{{ with . }}
onApply:
  backOffLimit: 1
  podSpec:
    tolerations: &tolerations {{.Tolerations | toJson }}
    nodeSelector: &node-selector {{.NodeSelector | toJson }}

    resources:
      requests:
        cpu: 500m
        memory: 1000Mi
      limits:
        cpu: 500m
        memory: 1000Mi

    containers:
      - name: main
        image: {{.JobImage}}
        imagePullPolicy: Always
        env:
          - name: KUBE_IN_CLUSTER_CONFIG
            value: "true"

          - name: KUBE_NAMESPACE
            value: {{.TfWorkspaceNamespace | squote}}
        command:
          - bash
          - -c
          - |+
            set -o pipefail
            set -o errexit

            eval $DECOMPRESS_CMD

            pushd "$TEMPLATES_DIR/{{.CloudProvider}}/work-machine"

            envsubst < state-backend.tf.tpl > state-backend.tf

            terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
            terraform workspace select --or-create {{.TFWorkspaceName}}

            cat > values.json <<'EOF'
            {{.ValuesJSON}}
            EOF

            terraform init -no-color 2>&1 | tee /dev/termination-log
            terraform plan -parallelism=2 --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
            terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log

            # terraform state pull | jq '.outputs' -r > outputs.json
            cat > secret.yml <<EOF
            apiVersion: v1
            kind: Secret
            metadata:
              name: {{.OutputSecretName}}
              namespace: {{.OutputSecretNamespace}}
            data: $(terraform state pull | jq '.outputs' -r -c)
            EOF

            kubectl apply -f secret.yml --server-side

    restartPolicy: Never

onDelete:
  backOffLimit: 1
  podSpec:
    tolerations: *tolerations
    nodeSelector: *node-selector

    resources:
      requests:
        cpu: 500m
        memory: 1000Mi
      limits:
        cpu: 500m
        memory: 1000Mi

    containers:
      - name: main
        image: {{.JobImage}}
        imagePullPolicy: Always
        env:
          - name: KUBE_IN_CLUSTER_CONFIG
            value: "true"

          - name: KUBE_NAMESPACE
            value: {{.TfWorkspaceNamespace | squote}}
        command:
          - bash
          - -c
          - |+
            set -o pipefail
            set -o errexit

            eval $DECOMPRESS_CMD

            pushd "$TEMPLATES_DIR/{{.CloudProvider}}/work-machine"

            envsubst < state-backend.tf.tpl > state-backend.tf

            terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
            terraform workspace select --or-create {{.TFWorkspaceName}}

            cat > values.json <<'EOF'
            {{.ValuesJSON}}
            EOF

            terraform init -no-color 2>&1 | tee /dev/termination-log
            terraform plan -parallelism=2 --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
            terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log

            kubectl delete secret/{{.OutputSecretName}} -n {{.OutputSecretNamespace}} --ignore-not-found=true
            
            kubectl delete node/{{.NodeName}}
    restartPolicy: Never

{{ end }}
