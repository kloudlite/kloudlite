{{ with . }}
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
        - name: mongodb
          image: &image mongo:latest
          imagePullPolicy: &pull-policy Always
          env: &env
            - name: MONGODB_URI
              valueFrom:
                secretKeyRef:
                  name: {{.RootCredentialsSecret}}
                  key: .CLUSTER_LOCAL_URI

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
            cat > /tmp/mongoscript.js <<EOF
            use $NEW_DB_NAME;

            if (db.getUser("$NEW_USERNAME") == null) {
              db.createUser({
                "user": "$NEW_USERNAME",
                "pwd": "$NEW_PASSWORD",
                "roles": [
                  {
                    "role": "readWrite",
                    "db": "$NEW_DB_NAME"
                  },
                  {
                    "role": "dbAdmin",
                    "db": "$NEW_DB_NAME"
                  }
                ]
              })
            }

            EOF
            
            echo connecting to "$MONGODB_URI"
            mongosh "$MONGODB_URI" < /tmp/mongoscript.js
      restartPolicy: Never

  onDelete:
    backOffLimit: 1
    podSpec:
      tolerations: *tolerations
      nodeSelector: *nodeselector

      resources: *resources

      containers:
        - name: mongodb
          image: *image
          imagePullPolicy: *pull-policy
          env: *env
          command:
            - sh
            - -c
            - |+
              cat > /tmp/mongoscript.js <<EOF
              use $NEW_DB_NAME;
              if (db.getUser("$NEW_USERNAME") != null) {
                db.dropUser("$NEW_USERNAME");
              }
              db.dropDatabase();
              EOF
              
              mongosh "$MONGODB_URI" < /tmp/mongoscript.js
      restartPolicy: Never
{{ end }}

