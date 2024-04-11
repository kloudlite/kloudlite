{{- if (eq .Values.cloudprovider "gcp") }}
apiVersion: batch/v1
kind: Job
metadata:
  name: gcp-csi-install
  namespace: {{.Release.Namespace}}
spec:
  template:
    spec:
      restartPolicy: Never
      securityContext:
        runAsUser: 0
      nodeSelector: {}
      tolerations: []
      serviceAccount: "{{.Values.serviceAccount.name}}"
      containers:
        - name: main
          {{- /* image: ghcr.io/kloudlite/kloudlite/operator/workers/helm-job-runner:v1.0.5-nightly */}}
          {{- /* image: gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine */}}
          {{- /* image: gcr.io/google.com/cloudsdktool/google-cloud-cli:slim */}}
          image: docker.io/nxtcoder17/k8s-utils:latest
          imagePullPolicy: Always
          env:
            - name: GCE_PD_SA_DIR
              value: /tmp

            - name: CLOUDSDK_AUTH_CREDENTIAL_FILE_OVERRIDE
              value: /tmp/cloud-sa.json

          command:
            - bash
            - -x
            - -c
            - |+
              set -o errexit
              set -o pipefail

              trap 'exit 0' SIGINT SIGTERM
              echo "hello"

              export GOPATH="/tmp"
              mkdir -p ${GOPATH}/src/sigs.k8s.io
              pushd ${GOPATH}/src/sigs.k8s.io

              git clone --depth 1 https://github.com/kubernetes-sigs/gcp-compute-persistent-disk-csi-driver 
              GCE_PD_DRIVER_VERSION=stable-master

              {{- /* ln -sf $(which bash) /bin/bash */}}
              {{- /* gcloud auth activate-service-account --key-file=/tmp/cloud-sa.json */}}
              {{- /**/}}
              {{- /* export PROJECT=$(grep -Po '"project_id": *\K"[^"]*"' "${GCE_PD_SA_DIR}/cloud-sa.json" | tr -d '"') */}}
              {{- /* GCE_PD_SA_NAME=gce-pd-csi-sa */}}

              pushd gcp-compute-persistent-disk-csi-driver

              {{- /* curl -LO https://dl.k8s.io/release/v1.29.2/bin/linux/amd64/kubectl > /usr/local/bin/kubectl && chmod +x /usr/local/bin/kubectl */}}
              {{- /**/}}
              {{- /* echo "################ RUNNING setup-project.sh ########################" */}}
              {{- /* bash ./gcp-compute-persistent-disk-csi-driver/deploy/setup-project.sh */}}
              {{- /* ./gcp-compute-persistent-disk-csi-driver/deploy/kubernetes/deploy-driver.sh */}}

              # assuming everything is setup correctly

              # sourced from ./gcp-compute-persistent-disk-csi-driver/deploy/kubernetes/deploy-driver.sh:98
              kubectl kustomize $PWD/deploy/kubernetes/overlays/${GCE_PD_DRIVER_VERSION} > /tmp/manifests.yml

              cat /tmp/manifests.yml | yq 'select(.kind != "Deployment")' | kubectl apply -f -

              cat /tmp/manifests.yml | yq 'select(.kind == "Deployment") | 
                .spec.template.spec.volumes = (
                  .spec.template.spec.volumes | map_values(
                      if .name == "cloud-sa-volume" then 
                        .secret.secretName = "example"
                      else 
                        .
                      end
                  )
                )' | kubectl apply -f -

              echo "installed"

          securityContext:
            privileged: true

          volumeMounts:
            - name: gcp-creds
              mountPath: /tmp/cloud-sa.json
              subPath: cloud-sa.json
              readOnly: true
      volumes:
        - name: "gcp-creds"
          secret:
            secretName: {{ include "gcp-credentials-secret-name" . }}
            items:
              - key: gcloud-creds.json
                path: cloud-sa.json
  backoffLimit: 1
{{- end }}
