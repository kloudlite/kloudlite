apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{.Values.crons.natsBackup.name}}
  namespace: {{.Release.Namespace}}
spec:
  schedule: "{{.Values.crons.natsBackup.schedule}}"
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            kloudlite.io/cron.for: "{{.Values.crons.natsBackup.name}}"
        spec:
          nodeSelector: {{.Values.crons.natsBackup.nodeSelector | toYaml | nindent 12}}
          tolerations: {{.Values.crons.natsBackup.tolerations | toYaml | nindent 12}}
          containers:
            - name: {{.Values.crons.natsBackup.name}}
              image: {{.Values.crons.natsBackup.image}}
              env:
                - name: BACKUP_DIR
                  value: &backup-dir "/backup"

                - name: NATS_URL
                  value: {{.Values.envVars.nats.url}}

                - name: NUM_BACKUPS
                  value: {{.Values.crons.natsBackup.numBackups | default 5 | squote}}

                - name: ENCRYPTION_PASSWORD
                  value: {{required ".values.crons.natsBackup.encryptionPassword is required" .Values.crons.natsBackup.encryptionPassword | squote}}
              volumeMounts:
                - mountPath: *backup-dir
                  name: backup
                  readOnly: false
          restartPolicy: OnFailure
          volumes:
            - name: backup
              persistentVolumeClaim:
                claimName: {{.Values.crons.natsBackup.name}}
