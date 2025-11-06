{{- $namespace := .Namespace }}
{{- $workmachineName := .WorkMachineName }}
{{- $sshUsername := .SSHUsername }}

{{- $sshUserUID := 1000 }}
{{- $sshUserGID := 1000 }}

apiVersion: v1
kind: Pod
metadata:
  name: hm-{{ $workmachineName }}
  namespace: {{ $namespace }}
  labels:
    app: workmachine-host-manager
    kloudlite.io/package-mgmt: "true"
    kloudlite.io/workmachine: {{ $workmachineName }}
spec:
  hostNetwork: true
  restartPolicy: Never
  dnsPolicy: None
  dnsConfig:
    nameservers:
      - 10.43.0.10
    searches:
      - {{ $namespace }}.svc.cluster.local
      - svc.cluster.local
      - cluster.local
    options:
      - name: ndots
        value: "5"
  serviceAccountName: workmachine-node-manager
  nodeName: {{ $workmachineName }}
  initContainers:
    - name: setup-nix
      image: ghcr.io/kloudlite/kloudlite/workmachine-node-manager:development
      imagePullPolicy: Always
      securityContext:
        privileged: true
      command: ["sh", "-c"]
      args:
        - |
          #!/bin/sh
          set -e

          echo "Checking if Nix is already on shared volume..."

          # Check if Nix is already on the shared volume
          if [ -f /nix-shared/var/nix/profiles/default/etc/profile.d/nix.sh ]; then
            echo "Nix already exists on shared volume, skipping copy"
          else
            echo "Copying Nix from image to shared volume..."

            # Copy the entire /nix directory from this container's image to the shared volume
            # The kloudlite/workmachine-node-manager image already has Nix installed at /nix
            # We need to copy it to the hostPath so it's available to other containers
            if [ -d /nix ]; then
              # Create target directory structure
              mkdir -p /nix-shared
              # Copy everything from /nix to /nix-shared
              cp -a /nix/* /nix-shared/
              echo "Nix copied successfully to shared volume"
            else
              echo "ERROR: /nix not found in image"
              exit 1
            fi
          fi

          # Always ensure profile directory exists (idempotent - safe to run multiple times)
          # This is required for nix-env to work properly with user profiles
          echo "Ensuring profile directory exists..."
          mkdir -p /nix-shared/profiles/per-user/root
          echo "Profile directory ready"
      volumeMounts:
        - name: nix-store
          mountPath: /nix-shared

        {{- /* - name: setup-ssh-key */}}
        {{- /*   image: busybox:latest */}}
        {{- /*   imagePullPolicy: IfNotPresent */}}
        {{- /*   command: */}}
        {{- /*     - sh */}}
        {{- /*     - -c */}}
        {{- /*     - cp /ssh-key-source/private-key /ssh-key-target/id_ed25519 && chown {{ $sshUserUID }}:{{ $sshUserGID }} /ssh-key-target/id_ed25519 && chmod 600 /ssh-key-target/id_ed25519 */}}
        {{- /*   volumeMounts: */}}
        {{- /*     - name: ssh-proxy-key */}}
        {{- /*       mountPath: /ssh-key-source */}}
        {{- /*       readOnly: true */}}
        {{- /*     - name: ssh-key-volume */}}
        {{- /*       mountPath: /ssh-key-target */}}

  containers:
    - name: workmachine-node-manager
      image: ghcr.io/kloudlite/kloudlite/workmachine-node-manager:development
      imagePullPolicy: Always
      securityContext:
        privileged: true
      env:
        - name: NAMESPACE
          value: {{ .Namespace }}
      volumeMounts:
        - name: nix-store
          mountPath: /nix
        - name: workspace-homes
          mountPath: /var/lib/kloudlite/workspace-homes
        - name: ssh-config
          mountPath: /var/lib/kloudlite/ssh-config

    - name: ssh-server
      image: linuxserver/openssh-server:latest
      imagePullPolicy: IfNotPresent
      env:
        - name: PUID
          value: "{{ $sshUserUID }}"
        - name: PGID
          value: "{{ $sshUserGID }}"
        - name: PASSWORD_ACCESS
          value: "false"
        - name: USER_PASSWORD
          value: kloudlite123
        - name: USER_NAME
          value: {{ $sshUsername }}
        - name: SUDO_ACCESS
          value: "false"
        - name: TCP_FORWARDING
          value: "true"

      ports:
        - name: ssh
          containerPort: 2222
          protocol: TCP

      volumeMounts:
        - name: ssh-key-volume
          mountPath: /config/.ssh/id_ed25519
          subPath: id_ed25519
          readOnly: true

        - name: ssh-config
          mountPath: /var/lib/kloudlite/ssh-config
          readOnly: false

        - name: sshd-config
          mountPath: /etc/ssh/sshd_config
          subPath: sshd_config
          readOnly: true

        - name: ssh-host-keys
          mountPath: /etc/ssh/ssh_host_rsa_key
          subPath: ssh_host_rsa_key
          readOnly: true
        - name: ssh-host-keys
          mountPath: /etc/ssh/ssh_host_rsa_key.pub
          subPath: ssh_host_rsa_key.pub
          readOnly: true
  volumes:
    - name: nix-store
      hostPath:
        path: /var/lib/kloudlite/nix-store
        type: DirectoryOrCreate
    - name: workspace-homes
      hostPath:
        path: /var/lib/kloudlite/workspace-homes
        type: DirectoryOrCreate
    - name: ssh-config
      hostPath:
        path: /var/lib/kloudlite/ssh-config
        type: DirectoryOrCreate
    - name: ssh-proxy-key
      secret:
        secretName: ssh-host-keys
    - name: ssh-key-volume
      emptyDir: {}
    - name: sshd-config
      configMap:
        name: sshd-config
    - name: ssh-host-keys
      secret:
        secretName: ssh-host-keys
        defaultMode: 0o600
