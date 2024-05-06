{{- /*gotype: github.com/kloudlite/operator/operators/nodepool/internal/templates.NodepoolJobVars*/ -}}
{{ with . }}
apiVersion: crds.kloudlite.io/v1
kind: Lifecycle
metadata: {{.JobMetadata | toYAML |nindent 2}}
spec:
  onApply:
    backOffLimit: 1
    podSpec:
      tolerations: &tolerations
        - effect: NoSchedule
          key: node-role.kubernetes.io/master
          operator: Exists

      nodeSelector: {{.NodeSelector | toYAML | nindent 10}}

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

              pushd "$TEMPLATES_DIR/{{.CloudProvider}}/worker-nodes"

              envsubst < state-backend.tf.tpl > state-backend.tf

              terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
              terraform workspace select --or-create {{.TFWorkspaceName}}

              cat > values.json <<'EOF'
              {{.ValuesJSON}}
              EOF

              terraform init -no-color 2>&1 | tee /dev/termination-log
              terraform plan -parallelism=2 --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log
      restartPolicy: Never

  onDelete:
    backOffLimit: 1
    podSpec:
      tolerations: *tolerations
      nodeSelector: {{.NodeSelector | toYAML | nindent 10}}

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

              pushd "$TEMPLATES_DIR/{{.CloudProvider}}/worker-nodes"

              envsubst < state-backend.tf.tpl > state-backend.tf

              terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
              terraform workspace select --or-create {{.TFWorkspaceName}}

              cat > values.json <<'EOF'
              {{.ValuesJSON}}
              EOF

              terraform init -no-color 2>&1 | tee /dev/termination-log
              terraform plan -parallelism=2 --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log
      restartPolicy: Never
{{ end }}
