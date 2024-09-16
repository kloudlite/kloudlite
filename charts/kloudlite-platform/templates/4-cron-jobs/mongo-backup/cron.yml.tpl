apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{.Values.crons.mongoBackup.name}}
  namespace: {{.Release.Namespace}}
spec:
  schedule: "{{.Values.crons.mongoBackup.schedule}}"
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            kloudlite.io/cron.for: "{{.Values.crons.mongoBackup.name}}"
        spec:
          nodeSelector: {{.Values.crons.mongoBackup.nodeSelector | toYaml | nindent 12}}
          tolerations: {{.Values.crons.mongoBackup.tolerations | toYaml | nindent 12}}
          containers:
            - name: {{.Values.crons.mongoBackup.name}}
              image: {{.Values.crons.mongoBackup.image}}
              env:
                - name: BACKUP_DIR
                  value: &backup-dir "/backup"

                - name: MONGODB_URI
                  valueFrom:
                    secretKeyRef:
                      name: "msvc-mongo-svc-creds"
                      key: .CLUSTER_LOCAL_URI

                - name: NUM_BACKUPS
                  value: {{.Values.crons.mongoBackup.numBackups | default 5 | squote}}

                - name: ENCRYPTION_PASSWORD
                  value: {{required ".values.crons.mongoBackup.encryptionPassword is required" .Values.crons.mongoBackup.encryptionPassword | squote}}
              volumeMounts:
                - mountPath: *backup-dir
                  name: backup
                  readOnly: false
          restartPolicy: OnFailure
          volumes:
            - name: backup
              persistentVolumeClaim:
                claimName: {{.Values.crons.mongoBackup.name}}
