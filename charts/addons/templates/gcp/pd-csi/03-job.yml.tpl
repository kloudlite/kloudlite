{{- if (and (eq .Values.cloudprovider "gcp") .Values.gcp.csi_driver.enabled )}}
apiVersion: crds.kloudlite.io/v1
kind: Job
metadata:
  name: gcp-csi
  namespace: {{include "gcp-csi-namespace" .}}
spec:
  onApply:
    backOffLimit: 1
    podSpec:
      tolerations:
        - operator: Exists
      containers:
        - name: main
          image: docker.io/nxtcoder17/k8s-utils:latest
          imagePullPolicy: Always
          command:
            - bash
            - -c
            - |+
              set -o errexit
              set -o pipefail

              trap 'exit 0' SIGINT SIGTERM
              echo "starting gcp csi installation"

              git clone --depth 1 https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver 

              pushd gcp-compute-persistent-disk-csi-driver

              kubectl kustomize $PWD/deploy/kubernetes/overlays/stable-master > /tmp/manifests.yml

              echo "Installing csi manifests"

              cat /tmp/manifests.yml | yq 'select(.kind != "Deployment")' | kubectl apply -f -

              cat /tmp/manifests.yml | yq --arg secret_name {{include "gcp-credentials-secret-name" .}} 'select(.kind == "Deployment") | 
                .spec.template.spec.volumes = (
                  .spec.template.spec.volumes | map_values(
                      if .name == "cloud-sa-volume" then 
                        .secret.secretName = $secret_name
                      else 
                        .
                      end
                  )
                )' | kubectl apply -f -


              # creating Storage Classes

              fstypes=("ext4" "xfs")
              for item in ${fstypes[@]}; do
              kubectl apply -f - <<EOF
              apiVersion: storage.k8s.io/v1
              kind: StorageClass
              metadata:
                name: sc-${item}
              parameters:
                csi.storage.k8s.io/fstype: ${item}
                type: pd-ssd
              provisioner: pd.csi.storage.gke.io
              reclaimPolicy: Delete
              volumeBindingMode: WaitForFirstConsumer
              allowVolumeExpansion: true
              EOF
              done

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
{{- end }}
