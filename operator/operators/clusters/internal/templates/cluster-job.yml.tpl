{{- /*gotype: github.com/kloudlite/operator/operators/clusters/internal/templates.ClusterJobVars*/ -}}
{{ with . }}
apiVersion: crds.kloudlite.io/v1
kind: Job
metadata: {{.JobMetadata | toYAML | nindent 4 }}
spec:
  onApply:
    backOffLimit: 1
    podSpec:
      restartPolicy: Never
      nodeSelector: {{ .NodeSelector | default dict | toYAML | nindent 8 }}
      tolerations: {{ .Tolerations | default list | toYAML | nindent 8 }}
      containers:
      - name: iac
        image: {{.JobImage}}
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
            value: {{ .TFWorkspaceSecretNamespace | squote}}

          {{- /* {{- if eq .CloudProvider "aws" }} */}}
          {{- /* - name: AWS_ACCESS_KEY_ID */}}
          {{- /*   value: {{.AWS.AccessKeyID}} */}}
          {{- /**/}}
          {{- /* - name: AWS_SECRET_ACCESS_KEY */}}
          {{- /*   value: {{.AWS.AccessKeySecret }} */}}
          {{- /* {{- end }} */}}

        command:
          - bash
          - -c
          - |+
            set -o pipefail
            set -o errexit

            eval $DECOMPRESS_CMD

            pushd "$TEMPLATES_DIR/{{.CloudProvider}}/master-nodes"
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
              name: {{.ClusterSecretName}}
              namespace: {{.ClusterSecretNamespace}}
            data:
              kubeconfig: $(cat outputs.json | jq '.kubeconfig.value')
              k3s_params: $(cat outputs.json | jq -r '."kloudlite-k3s-params".value' | base64 | tr -d '\n')
              k3s_agent_token: $(cat outputs.json | jq -r '.k3s_agent_token.value' | base64 | tr -d '\n')
            EOF
            exit 0

  onDelete:
    backOffLimit: 1
    podSpec:
      restartPolicy: Never
      containers:
        - name: iac
          image: {{.JobImage}}
          imagePullPolicy: "Always"

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
              value: {{.TFWorkspaceSecretNamespace | squote}}

          command:
            - bash
            - -c
            - |+
              set -o pipefail
              set -o errexit

              eval $DECOMPRESS_CMD

              pushd "$TEMPLATES_DIR/{{.CloudProvider}}/master-nodes"

              envsubst < state-backend.tf.tpl > state-backend.tf

              terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log
              terraform workspace select --or-create {{.TFWorkspaceName}}

              cat > values.json <<EOF
              {{.ValuesJSON}}
              EOF

              terraform plan -parallelism=2 --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log
              terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log
              kubectl delete secret/{{.ClusterSecretName}} -n {{.ClusterSecretNamespace}} --ignore-not-found=true
              exit 0

{{ end }}
