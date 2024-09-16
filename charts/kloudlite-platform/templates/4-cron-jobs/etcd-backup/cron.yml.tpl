apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{.Values.crons.etcdBackup.name}}
  namespace: {{.Release.Namespace}}
spec:
  schedule: "{{.Values.crons.etcdBackup.schedule}}"
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            kloudlite.io/cron.for: "{{.Values.crons.etcdBackup.name}}"
        spec:
          nodeSelector: {{.Values.crons.etcdBackup.nodeSelector | toYaml | nindent 12}}
          tolerations: {{.Values.crons.etcdBackup.tolerations | toYaml | nindent 12}}
          containers:
            - name: {{.Values.crons.etcdBackup.name}}
              image: {{.Values.crons.etcdBackup.image}}
              env:
                - name: BACKUP_DIR
                  value: &backup-dir "/backup"

                - name: SNAPSHOTS_DIR
                  value: &snapshots-dir "/etcd-snapshots"

                - name: NUM_BACKUPS
                  value: {{.Values.crons.etcdBackup.numBackups | default 5 | squote}}

                - name: ENCRYPTION_PASSWORD
                  value: {{required ".values.crons.etcdBackup.encryptionPassword is required" .Values.crons.etcdBackup.encryptionPassword | squote}}
              volumeMounts:
                - mountPath: *backup-dir
                  name: backup
                  readOnly: false
                - mountPath: *snapshots-dir
                  name: etcd-snapshots
                  readOnly: true
          restartPolicy: OnFailure
          volumes:
            - name: backup
              persistentVolumeClaim:
                claimName: {{.Values.crons.etcdBackup.name}}

            - name: etcd-snapshots
              hostPath:
                path: "/var/lib/rancher/k3s/server/db/snapshots" # k3s etcd snapshots dir
                type: Directory
