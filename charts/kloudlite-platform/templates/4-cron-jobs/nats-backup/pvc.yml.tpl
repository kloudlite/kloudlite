# [source](https://github.com/yandex-cloud/k8s-csi-s3/blob/master/deploy/kubernetes/examples/pvc-manual.yaml)
# it is required to create a statically provisioned PVC, so that it won't be removed when you remove the PVC/PV

apiVersion: v1
kind: PersistentVolume
metadata:
  name: {{.Values.crons.natsBackup.name}}
  namespace: {{.Release.Namespace}}
spec:
  storageClassName: "{{.Values.csiS3.storageClass}}"
  capacity:
    storage: "{{.Values.crons.natsBackup.storageSize}}"
  accessModes:
    - ReadWriteMany
  claimRef:
    namespace: "{{.Release.Namespace}}"
    name: "{{.Values.crons.natsBackup.name}}"
  csi:
    driver: "{{.Values.csiS3.driver}}"
    volumeAttributes:
      capacity: "{{.Values.crons.natsBackup.storageSize}}"
      mounter: geesefs
      options: --memory-limit 1000 --dir-mode 0777 --file-mode 0666

    volumeHandle: {{.Values.csiS3.s3.bucketName}}/{{.Values.crons.natsBackup.name}}

    controllerPublishSecretRef: &csi-s3-secret
      name: {{.Values.csiS3.storageClass}}-secret
      namespace: {{.Release.Namespace}}
    nodePublishSecretRef: *csi-s3-secret
    nodeStageSecretRef: *csi-s3-secret

---

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: "{{.Values.crons.natsBackup.name}}"
  namespace: {{.Release.Namespace}}
  annotations:
    volume.beta.kubernetes.io/storage-provisioner: {{.Values.csiS3.driver}}
    volume.kubernetes.io/storage-provisioner: {{.Values.csiS3.driver}}
spec:
  # Empty storage class disables dynamic provisioning
  storageClassName: ""
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: "{{.Values.crons.natsBackup.storageSize}}"
