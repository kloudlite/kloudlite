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
          image: &image postgres:13
          imagePullPolicy: &pull-policy IfNotPresent
          env: &env
            - name: POSTGRES_URI
              valueFrom:
                secretKeyRef:
                  name: {{.PostgressRootCredentialsSecret}}
                  key: .CLUSTER_LOCAL_URI

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

  onDelete:
    backOffLimit: 1
    podSpec:
      tolerations: *tolerations
      nodeSelector: *nodeselector
      resources: *resources

      containers:
        - name: postgres
          image: *image
          imagePullPolicy: *pull-policy
          env: *env
          command:
            - sh
            - -c
            - |+
              cat > script.sql <<EOF
              DROP DATABASE $NEW_DB_NAME;
              DROP USER $NEW_USERNAME;
              EOF
              
              psql "$POSTGRES_URI" -v ON_ERROR_STOP=1 -f ./script.sql
      restartPolicy: Never
{{ end }}

