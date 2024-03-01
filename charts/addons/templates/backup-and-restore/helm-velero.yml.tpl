{{- if .Values.common.velero.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: velero
  namespace: {{.Release.Namespace}}
spec:
  chartRepoURL: https://vmware-tanzu.github.io/helm-charts
  chartName: "velero"
  chartVersion: 5.4.1
  jobVars:
    tolerations:
      - operator: Exists
  values:
    nameOverride: "velero"
    {{- /* fullnameOverride: "velero" */}}
    initContainers:
      - name: velero-plugin-for-csi
        image: velero/velero-plugin-for-csi:v0.7.0
        imagePullPolicy: IfNotPresent
        volumeMounts:
          - mountPath: /target
            name: plugins

      {{- if (eq .Values.cloudprovider "aws") }}
      - name: velero-plugin-for-aws
        image: velero/velero-plugin-for-aws:v1.9.0
        imagePullPolicy: IfNotPresent
        volumeMounts:
          - mountPath: /target
            name: plugins
      {{- end }}

    nodeSelector: {{include "node-selector-masters" . | nindent 6 }}
    tolerations: {{include "node-tolerations-masters" . | nindent 6 }}
    priorityClassName: {{.Values.statefulPriorityClassName}}

    credentials:
      useSecret: {{.Values.common.velero.configuration.useS3Credentials.enabled }}
      name: "velero-s3-credentials"
      secretContents:
        cloud: |
          [default]
          aws_access_key_id={{required .Values.common.velero.configuration.useS3Credentials.creds.accessKey ".Values.common.velero.configuration.useS3Credentials.creds.accessKey must be provided" }}
          aws_secret_access_key={{required .Values.common.velero.configuration.useS3Credentials.creds.secretKey ".Values.velero.configuration.s3Credentials.creds.secretKey must be provided" }}

    deployNodeAgent: true
    nodeAgent:
      tolerations:
        - operator: Exists

    configuration:
      features: EnableCSI
      backupStorageLocation:
        - name: default
          provider: "{{.Values.cloudprovider }}"
          bucket: {{.Values.common.velero.configuration.backupStorage.bucket}}
          region: {{.Values.common.velero.configuration.backupStorage.region }}
          s3ForcePathStyle: true

          {{- if .Values.common.velero.configuration.backupStorage.path }}
          prefix: {{.Values.common.velero.configuration.backupStorage.path}}
          {{- end }}

          {{- if .Values.common.velero.configuration.backupStorage.s3Url }}
          s3Url: {{.Values.common.velero.configuration.backupStorage.s3Url | squote}}
          {{- end }}
          config:
            region: "{{.Values.common.velero.configuration.backupStorage.region }}"

      volumeSnapshotLocation:
        - name: default
          provider: "{{ .Values.cloudprovider }}"
          config:
            region: "{{.Values.common.velero.configuration.backupStorage.region }}"
            incremental: true
{{- end }}
