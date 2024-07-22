{{- $cronName := "mongo-csi-s3-backup" }}

apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{$cronName}}
  namespace: {{.Release.Namespace}}
spec:
  schedule: "{{.Values.crons.mongoBackup.configuration.schedule}}"
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            app: {{$cronName}}
        spec:
          containers:
            - name: {{$cronName}}
              image: {{.Values.crons.mongoBackup.configuration.image}}
              command: 
                - /bin/sh
                - -c
                - |
                  apk add zip
                  set -o errexit
                  set -o pipefail

                  trap 'echo "Backup failed"; exit 1' ERR

                  BACKUP_DEST="/mongo-backups"

                  TIMESTAMP=$(date +"%Y%m%d%H%M%S")
                  FILENAME="backup_${TIMESTAMP}"
                  BACKUP_DIR="${BACKUP_DEST}/${FILENAME}.zip"

                  BACKUP_TEMP_DIR="/tmp/${FILENAME}"

                  mkdir -p "$BACKUP_DEST"
                  mongodump --uri ${MONGODB_URI} --archive=${BACKUP_TEMP_DIR} --dumpDbUsersAndRoles --gzip

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
                - name: MONGODB_URI
                  value: {{.Values.crons.mongoBackup.configuration.mongodbUri}}
                - name: NUM_BACKUPS
                  value: {{.Values.crons.mongoBackup.configuration.numBackups | default "5"}}
                - name: ENCRYPTION_PASSWORD
                  value: {{.Values.crons.mongoBackup.configuration.encryptionPassword}}
              volumeMounts:
                - mountPath: /mongo-backups
                  name: mongo-backups
          restartPolicy: OnFailure
          volumes:
            - name: mongo-backups
              persistentVolumeClaim:
                claimName: {{$cronName}}
                readOnly: false
