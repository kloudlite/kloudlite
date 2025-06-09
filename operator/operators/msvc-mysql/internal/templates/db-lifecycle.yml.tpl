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
        - name: mysql
          image: &image ghcr.io/kloudlite/hub/mysql-client:latest
          imagePullPolicy: &pull-policy Always
          env: &env
            - name: MYSQL_HOST
              valueFrom:
                secretKeyRef:
                  name: {{.RootCredentialsSecret}}
                  key: .CLUSTER_LOCAL_HOST

            - name: MYSQL_USERNAME
              valueFrom:
                secretKeyRef:
                  name: {{.RootCredentialsSecret}}
                  key: USERNAME

            - name: MYSQL_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{.RootCredentialsSecret}}
                  key: PASSWORD

            - name: NEW_DB_NAME
              valueFrom:
                secretKeyRef:
                  name: {{.NewCredentialsSecret}}
                  key: DB_NAME

            - name: NEW_USERNAME
              valueFrom:
                secretKeyRef:
                  name: {{.NewCredentialsSecret}}
                  key: USERNAME

            - name: NEW_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{.NewCredentialsSecret}}
                  key: PASSWORD

          command:
          - sh
          - -c
          - |+
            cat > /tmp/script.sql <<EOF
            CREATE DATABASE IF NOT EXISTS $NEW_DB_NAME;
            CREATE USER IF NOT EXISTS '$NEW_USERNAME'@'%%' IDENTIFIED BY '$NEW_PASSWORD';

            GRANT ALL PRIVILEGES ON $NEW_DB_NAME.* TO '$NEW_USERNAME'@'%%';
            FLUSH PRIVILEGES;
            EOF

            mysql -h "$MYSQL_HOST" -u "$MYSQL_USERNAME" -p"$MYSQL_PASSWORD" < /tmp/script.sql
      restartPolicy: Never

  onDelete:
    backOffLimit: 1
    podSpec:
      tolerations: *tolerations
      nodeSelector: *nodeselector
      resources: *resources

      containers:
        - name: mysql
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
              
              mysql -h "$MYSQL_HOST" -u "$MYSQL_USERNAME" -p"$MYSQL_PASSWORD" < /tmp/script.sql
      restartPolicy: Never

{{ end }}

