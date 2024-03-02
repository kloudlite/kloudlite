{{- if (eq .Values.cloudprovider "aws") }}
{{- if .Values.aws.ebs_csi_driver.enabled }}
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: aws-ebs-csi
  namespace: kube-system
spec:
  chartRepoURL: https://kubernetes-sigs.github.io/aws-ebs-csi-driver
  chartVersion: 2.22.0
  chartName: aws-ebs-csi-driver

  jobVars:
    tolerations:
      - operator: Exists

  preInstall: |+
    # volume snapshot classes
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml

    # volume snapshot contents
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml

    # volume snapshots
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml

    # installing volume snapshot controller RBACs
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/deploy/kubernetes/snapshot-controller/rbac-snapshot-controller.yaml

    # installing volume snapshot controller deployment
    kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/deploy/kubernetes/snapshot-controller/setup-snapshot-controller.yaml

  postInstall: |+
    echo "making sure sc-ext4 is the default storage class"

    kubectl get sc/local-path -o=jsonpath={.metadata.name}
    exit_code=$?

    if [ $exit_code -eq 0 ]; then
      kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
    fi

    kubectl get sc/sc-ext4 -o=jsonpath={.metadata.name}
    exit_code=$?
    if [ $exit_code -eq 0 ]; then
      kubectl patch storageclass sc-ext4 -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
    fi

    {{- /* kubectl apply -f - <<EOF */}}
    {{- /* apiVersion: batch/v1 */}}
    {{- /* kind: Job */}}
    {{- /* metadata: */}}
    {{- /*   name: ensure-default-storage-class */}}
    {{- /*   namespace: {{ .Release.Namespace }} */}}
    {{- /* spec: */}}
    {{- /*   template: */}}
    {{- /*     spec: */}}
    {{- /*       tolerations: */}}
    {{- /*         - operator: Exists */}}
    {{- /*       serviceAccountName: {{.Values.serviceAccount.name}} */}}
    {{- /*       containers: */}}
    {{- /*       - name: kubectl */}}
    {{- /*         image: bitnami/kubectl:latest */}}
    {{- /*         command: ["sh"] */}}
    {{- /*         args: */}}
    {{- /*         - -c */}}
    {{- /*         - |+ #bash */}}
    {{- /*           kubectl get sc/local-path -o=jsonpath={.metadata.name} */}}
    {{- /*           exit_code=$? */}}
    {{- /**/}}
    {{- /*           if [ $exit_code -eq 0 ]; then */}}
    {{- /*             kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}' */}}
    {{- /*           fi */}}
    {{- /**/}}
    {{- /*           kubectl get sc/sc-ext4 -o=jsonpath={.metadata.name} */}}
    {{- /*           exit_code=$? */}}
    {{- /*           if [ $exit_code -eq 0 ]; then */}}
    {{- /*             kubectl patch storageclass sc-ext4 -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}' */}}
    {{- /*           fi */}}
    {{- /*       restartPolicy: Never */}}
    {{- /*   backoffLimit: 0 */}}
    {{- /* EOF */}}

  values:
    customLabels:
      kloudlite.io/part-of: "{{.Chart.Name}}"
    storageClasses: 
      - name: sc-xfs
        labels:
          kloudlite.io/part-of: {{.Chart.Name}}
        volumeBindingMode: "WaitForFirstConsumer"
        reclaimPolicy: "Retain"
        parameters:
          encrypted: "false"
          type: "gp3"
          fsType: "xfs"

      - name: sc-ext4
        labels:
          kloudlite.io/part-of: {{.Chart.Name}}
        volumeBindingMode: "WaitForFirstConsumer"
        reclaimPolicy: "Retain"
        parameters:
          encrypted: "false"
          type: "gp3"
          fsType: "ext4"
    controller:
      nodeSelector: {{include "node-selector-masters" . | nindent 8 }}
      tolerations: {{include "node-tolerations-masters" . | nindent 8 }}
    node:
      nodeSelector: {{include "node-selector-agent" . | nindent 8 }}

      # tolerate any taints
      tolerations:
        - operator: "Exists"
{{- end }}
{{- end }}
