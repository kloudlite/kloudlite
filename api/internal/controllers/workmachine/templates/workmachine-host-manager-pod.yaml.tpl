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
  serviceAccountName: hm-{{ $workmachineName }}
  nodeName: {{ $workmachineName }}
  hostPID: true
  dnsConfig:
    searches:
      - {{ .TargetNamespace }}.svc.cluster.local
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

  containers:
    - name: host-manager
      image: {{ .HostManagerImage }}
      imagePullPolicy: Always
      securityContext:
        privileged: true
      ports:
        - name: metrics
          containerPort: 8081
          protocol: TCP
      env:
        - name: NAMESPACE
          value: {{ .Namespace }}
        - name: WORKMACHINE_NAME
          value: {{ $workmachineName }}
      volumeMounts:
        - name: nix-store
          mountPath: /nix
        - name: workspace-homes
          mountPath: /var/lib/kloudlite/home
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
