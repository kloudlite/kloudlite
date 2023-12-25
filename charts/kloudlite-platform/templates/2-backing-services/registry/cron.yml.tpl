{{- $configName := .Release.Name  -}}
{{- $name := .Release.Name -}}

apiVersion: batch/v1
kind: CronJob
metadata:
  name: {{ $name }}
spec:
  schedule: "0 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
          - name: registry-container
            image: registry:2
            command:
            - /bin/sh
            - -c
            -  registry garbage-collect --dry-run /etc/docker/registry/config.yml 
            # --delete-untagged=true
            imagePullPolicy: IfNotPresent
            resources:
              limits:
                cpu: 100m
                memory: 100Mi
              requests:
                cpu: 100m
                memory: 100Mi
            volumeMounts:
            - name: config-volume
              mountPath: /etc/docker/registry

          volumes:
          - name: config-volume
            configMap:
              name: {{ $configName }}
