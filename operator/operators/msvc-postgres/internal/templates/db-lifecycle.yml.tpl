{{ with . }}
{{- /* apiVersion: crds.kloudlite.io/v1 */}}
{{- /* kind: Lifecycle */}}
{{- /* metadata: {{.Metadata | toYAML |nindent 2}} */}}
spec:
  onApply:
    backOffLimit: 1
    podSpec:
      tolerations: &tolerations {{.Tolerations | toYAML | nindent 10}}
      nodeSelector: &nodeselector {{.NodeSelector | toYAML | nindent 10}}

      resources: &resources
        requests:
          cpu: 500m
          memory: 1000Mi
        limits:
          cpu: 500m
          memory: 1000Mi

      containers:
        - name: postgres
          image: postgres:13
          env:
            - name: POSTGRES_URI
              valueFrom:
                secretKeyRef:
                  name: {{.PostgressRootCredentialsSecret}}
                  key: CLUSTER_LOCAL_URI

            - name: NEW_DB_NAME
              valueFrom:
                secretKeyRef:
                  name: {{.PostgressNewCredentialsSecret}}
                  key: DB_NAME

            - name: NEW_USERNAME
              valueFrom:
                secretKeyRef:
                  name: {{.PostgressNewCredentialsSecret}}
                  key: USERNAME

            - name: NEW_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{.PostgressNewCredentialsSecret}}
                  key: PASSWORD

          command:
          - sh
          - -c
          - |+
            cat > script.sql <<EOF
            CREATE DATABASE $NEW_DB_NAME;
            CREATE USER $NEW_USERNAME WITH ENCRYPTED PASSWORD '$NEW_PASSWORD';

            GRANT ALL ON DATABASE $NEW_DB_NAME TO $NEW_USERNAME;
            ALTER DATABASE $NEW_DB_NAME OWNER TO $NEW_USERNAME;

            EOF

            psql "$POSTGRES_URI" -v ON_ERROR_STOP=1 -f ./script.sql
      restartPolicy: Never

  {{- /* onDelete: */}}
  {{- /*   backOffLimit: 1 */}}
  {{- /*   podSpec: */}}
  {{- /*     tolerations: *tolerations */}}
  {{- /*     nodeSelector: *nodeselector */}}
  {{- /**/}}
  {{- /*     resources: *resources */}}
  {{- /**/}}
  {{- /*     containers: */}}
  {{- /*       - name: main */}}
  {{- /*         image: {{.JobImage}} */}}
  {{- /*         imagePullPolicy: Always */}}
  {{- /*         env: */}}
  {{- /*           - name: KUBE_IN_CLUSTER_CONFIG */}}
  {{- /*             value: "true" */}}
  {{- /**/}}
  {{- /*           - name: KUBE_NAMESPACE */}}
  {{- /*             value: {{.TfWorkspaceNamespace | squote}} */}}
  {{- /*         command: */}}
  {{- /*           - bash */}}
  {{- /*           - -c */}}
  {{- /*           - |+ */}}
  {{- /*             set -o pipefail */}}
  {{- /*             set -o errexit */}}
  {{- /**/}}
  {{- /*             eval $DECOMPRESS_CMD */}}
  {{- /**/}}
  {{- /*             pushd "$TEMPLATES_DIR/{{.CloudProvider}}/worker-nodes" */}}
  {{- /**/}}
  {{- /*             envsubst < state-backend.tf.tpl > state-backend.tf */}}
  {{- /**/}}
  {{- /*             terraform init -reconfigure -no-color 2>&1 | tee /dev/termination-log */}}
  {{- /*             terraform workspace select --or-create {{.TFWorkspaceName}} */}}
  {{- /**/}}
  {{- /*             cat > values.json <<'EOF' */}}
  {{- /*             {{.ValuesJSON}} */}}
  {{- /*             EOF */}}
  {{- /**/}}
  {{- /*             terraform init -no-color 2>&1 | tee /dev/termination-log */}}
  {{- /*             terraform plan -parallelism=2 --destroy --var-file ./values.json -out=tfplan -no-color 2>&1 | tee /dev/termination-log */}}
  {{- /*             terraform apply -parallelism=2 -no-color tfplan 2>&1 | tee /dev/termination-log */}}
  {{- /*     restartPolicy: Never */}}
{{ end }}

