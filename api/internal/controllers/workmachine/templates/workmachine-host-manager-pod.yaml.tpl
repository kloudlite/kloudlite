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
  restartPolicy: Never
  serviceAccountName: workmachine-node-manager
  nodeName: {{ $workmachineName }}
  dnsConfig:
    searches:
      - {{ .TargetNamespace }}.svc.cluster.local
  initContainers:
    - name: setup-nix
      image: ghcr.io/kloudlite/kloudlite/workmachine-node-manager:v1.0.9-nightly
      imagePullPolicy: IfNotPresent
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
            # Always ensure profile directory exists (idempotent - safe to run multiple times)
            echo "Ensuring profile directory exists..."
            mkdir -p /nix-shared/profiles/per-user/root
            echo "Profile directory ready"
            exit 0
          fi

          echo "Copying Nix from image to shared volume..."

          # Verify Nix exists in the image
          if [ ! -d /nix ]; then
            echo "ERROR: /nix not found in image"
            exit 1
          fi

          # Create shared volume directory
          mkdir -p /nix-shared

          # Copy Nix from image to shared volume
          cp -a /nix/* /nix-shared/

          echo "Nix copied successfully to shared volume"

          # Ensure profile directory exists
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
      image: {{ .HostManagerImage }}
      imagePullPolicy: Always
      command: ["/usr/local/bin/workmachine-node-manager"]
      securityContext:
        privileged: true
      env:
        - name: NAMESPACE
          value: {{ .Namespace }}
        - name: WORKMACHINE_NAME
          value: {{ $workmachineName }}
      volumeMounts:
        - name: nix-store
          mountPath: /nix
        - name: workspace-homes
          mountPath: /var/lib/kloudlite/workspace-homes
        - name: ssh-config
          mountPath: /var/lib/kloudlite/ssh-config
        # Host filesystem mounts for GPU detection and driver installation
        - name: host-sys
          mountPath: /host/sys
          readOnly: true
        - name: host-dev
          mountPath: /host/dev
        - name: host-proc
          mountPath: /host/proc
          readOnly: true
        - name: host-lib-modules
          mountPath: /host/lib/modules

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
    # Host filesystem volumes for GPU detection and driver installation
    - name: host-sys
      hostPath:
        path: /sys
        type: Directory
    - name: host-dev
      hostPath:
        path: /dev
        type: Directory
    - name: host-proc
      hostPath:
        path: /proc
        type: Directory
    - name: host-lib-modules
      hostPath:
        path: /lib/modules
        type: DirectoryOrCreate
    - name: ssh-proxy-key
      secret:
        secretName: ssh-host-keys-{{.WorkMachineName}}
    - name: ssh-key-volume
      emptyDir: {}
    - name: sshd-config
      configMap:
        name: sshd-config
    - name: ssh-host-keys
      secret:
        secretName: ssh-host-keys-{{.WorkMachineName}}
        defaultMode: 0o600
