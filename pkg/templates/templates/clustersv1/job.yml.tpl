{{- $name := get . "name"}}
{{- $namespace := get . "namespace"}}
{{- $ownerRefs := get . "ownerRefs" }}

{{- $cludProvider := get . "cloudProvider"}}
{{- $action := get . "action"}}
{{- $nodeConfig := get . "nodeConfig"}}

{{- $providerConfig := get . "providerConfig"}}
{{- $AwsProvider := get . "AwsProvider"}}
{{- $AzureProvider := get . "AzureProvider"}}
{{- $DoProvider := get . "DoProvider"}}
{{- $GCPProvider := get . "GCPProvider"}}


apiVersion: batch/v1
kind: Job
metadata:
  name: {{$name}}
  namespace: {{$namespace}}
  ownerReferences: {{$ownerRefs | toJson}}
spec:
  template:
    spec:
      # nodeSelector:
      #   kloudlite.io/auto-scaler: "true"
      tolerations:
      - effect: NoExecute
        key: kloudlite.io/auto-scaler
        operator: Exists

      restartPolicy: Never
      # serviceAccount: cluster-kloudlite-svc-account
      serviceAccountName: kloudlite-cluster-svc-account
      # nodeSelector:
      #   kloudlite.io/region: kl-blr1
      containers:

      - image: registry.kloudlite.io/kloudlite/development/nodecontroller:v1.0.5
        name: nodectrl
        imagePullPolicy: Always
        command:
          - /bin/sh
          - -c
          - |+
            trap 'touch /usr/share/pod/done' EXIT

            while [ ! -f "$S3_DIR/checkpoint" ] 
            do
              sleep 2
            done
            # mkdir -p /home/nonroot/ssh
            # cat /ssh/access.pub > /home/nonroot/ssh/access.pub
            # cat /ssh/id_rsa > /home/nonroot/ssh/id_rsa
            # chown -R nonroot:nonroot /home/nonroot/ssh
            # chmod 400 /home/nonroot/ssh/id_rsa
            # tail -f /dev/null
            ./nodecontroller

        # securityContext:
        #   runAsNonRoot: true
        #   runAsUser: 1000
        resources: {} #  needed to add after inspection
        env:
        - name: KL_CONFIG
          value: {{ $klConfig }}
        - name: NODE_CONFIG
          value: {{ $nodeConfig }}
        - name: PROVIDER
          value: {{ $provider }}
        - name: S3_DIR
          value: /terraform/storage

        volumeMounts:

          - mountPath: /terraform/storage
            mountPropagation: HostToContainer
            name: shared-data

          - mountPath: /usr/share/pod
            name: tmp-pod

          - name: agent-ssh
            mountPath: "/home/nonroot/ssh"


      - image: nxtcoder17/s3fs-mount:v1.0.0
        name: spaces-sidecar
        envFrom:
          - secretRef:
              name: s3-secret
          - configMapRef:
              name: s3-config
        env:
          - name: MOUNT_DIR
            value: "/data"
          - name: "BUCKET_DIR"
            value: /terraform/storage  # mount example-bucket/images
        imagePullPolicy: Always
        command:
          - bash
          - -c
          - |+
            chown -R 1000:1000 $MOUNT_DIR
            bash run.sh &
            while ! test -f /usr/share/pod/done; do
              # echo 'Waiting for the agent pod to finish...'
              sleep 5
            done
            # echo "Agent pod finished, exiting"
            exit 0
        resources:
          requests:
            cpu: 150m
            memory: 150Mi
          limits:
            cpu: 200m
            memory: 200Mi
        securityContext:
          capabilities:
            add:
            - SYS_ADMIN
          privileged: true
        volumeMounts:
        - mountPath: /data
          mountPropagation: Bidirectional
          name: shared-data

        - mountPath: /usr/share/pod
          name: tmp-pod

      volumes:
        - emptyDir: {}
          name: shared-data

        - emptyDir: {}
          name: tmp-pod

        - name: agent-ssh
          secret:
            secretName: {{ $sshSecretName }}
            # optional: false # default setting; "mysecret" must exist

