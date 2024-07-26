apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{.Values.crons.natsBackup.name}}
  namespace: {{.Release.Namespace}}
spec:
  schedule: "{{.Values.crons.natsBackup.configuration.schedule}}"
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            app: {{.Values.crons.natsBackup.name}}
        spec:
          containers:
            - name: {{.Values.crons.natsBackup.name}}
              image: {{.Values.crons.natsBackup.configuration.image}}
              command: 
                - sh
                - -c
                - |
                  set -o errexit
                  set -o pipefail

                  apk add zip

                  trap 'echo "Backup failed"; exit 1' ERR

                  BACKUP_DEST="/nats-backups"

                  TIMESTAMP=$(date +"%Y_%m_%d_%H_%M_%S")
                  FILENAME="nats_backup_${TIMESTAMP}"
                  BACKUP_DIR="${BACKUP_DEST}/${FILENAME}.zip"

                  BACKUP_TEMP_DIR="/tmp/${FILENAME}"

                  mkdir -p "$BACKUP_DEST"
                  nats account backup --server=$NATS_URL "$BACKUP_TEMP_DIR" -f

                  zip -r -P "$ENCRYPTION_PASSWORD" "${BACKUP_TEMP_DIR}.zip" "$BACKUP_TEMP_DIR"

                  cp "${BACKUP_TEMP_DIR}.zip" "$BACKUP_DIR"

                  cd "$BACKUP_DEST" || exit
                  BACKUPS=$(ls -1t)

                  COUNT=0
                  for BACKUP in $BACKUPS; do
                    echo "Processing backup $BACKUP"
                    COUNT=$((COUNT + 1))
                    if [ "$COUNT" -gt "$NUM_BACKUPS" ]; then
                      rm -rf "$BACKUP"
                    fi
                  done

                  echo "Backup completed"
              env:
                - name: NATS_URL
                  {{- /* value: {{.Values.crons.natsBackup.configuration.server}} */}}
                  value: {{.Values.envVars.nats.url}}

                - name: NUM_BACKUPS
                  value: {{.Values.crons.natsBackup.configuration.numBackups | default 5 | squote}}
                - name: ENCRYPTION_PASSWORD
                  value: {{.Values.crons.natsBackup.configuration.encryptionPassword}}
              volumeMounts:
                - mountPath: /nats-backups
                  name: nats-backups
          restartPolicy: OnFailure
          volumes:
            - name: nats-backups
              persistentVolumeClaim:
                claimName: {{.Values.crons.natsBackup.name}}
                readOnly: false
