{{- with . }}
apiVersion: apps/v1
kind: StatefulSet
metadata: {{.Metadata | toJson }}
spec:
  replicas: {{.Paused | ternary 0 1 }}
  selector:
    matchLabels: {{ .Selector | toJson }}
  template:
    metadata:
      labels: 
        {{- range $k, $v := .Selector }}
        {{$k}}: {{$v}}
        {{- end }}
        {{- range $k, $v := .PodLabels }}
        {{$k}}: {{$v}}
        {{- end }}
    spec:
      securityContext:
        fsGroup: 1000
      hostname: {{.Metadata.Name}}
      nodeName: {{.WorkMachineName}}

      {{- if .ServiceAccountName }}
      serviceAccount: {{.ServiceAccountName | squote}}
      {{- end }}

      tolerations:
        - key: {{.WorkMachineTolerationKey}}
          operator: "Equal"
          value: {{.WorkMachineName |squote}}
          effect: "NoExecute"

      initContainers:
        - name: volume-permissions
          image: {{.ImageInitContainer}}
          imagePullPolicy: Always
          env:
            - name: KL_WORKSPACE
              value: {{.Metadata.Name}}

            - name: HOME
              value: "/home/kl"

            - name: KL_BOX_MODE
              value: "true"
          command:
            - "bash"
            - "-c"
            - |
              set -e
              set +x

              chown -R kl /nix

              # creates ssh directory
              mkdir -p $HOME/workspaces
              chown -R kl $HOME

          volumeMounts: &volume-mounts
            - mountPath: /home/kl
              name: home-dir
            
            - mountPath: /nix
              name: nix-dir

        - name: init-home-dir
          image: {{.ImageInitContainer}}
          imagePullPolicy: Always
          env:
            - name: KL_WORKSPACE
              value: {{.Metadata.Name}}

            - name: HOME
              value: "/home/kl"

            - name: KL_BOX_MODE
              value: "true"
          securityContext:
            runAsUser: 1000
            runAsGroup: 1000
          command:
          - "bash"
          - "-c"
          - |
            set -e
            set +x

            mkdir -p $HOME/.ssh
            cp /tmp/authorized_keys /tmp/id_rsa.pub /tmp/id_rsa $HOME/.ssh/

            export PATH="$HOME/.nix-profile/bin:$HOME/.local/bin:$PATH"

            if ! command -v nix >/dev/null 2>&1; then
              echo "[#] nix is not installed, will install it"
              curl -L https://nixos.org/nix/install | sh
            fi

            append_if_not_exists() {
              line=$1
              file=$2

              mkdir -p "$(dirname $file)"
              grep -qxF "$line" "$file" || echo "$line" >> "$file"
            }

            append_if_not_exists 'experimental-features = nix-command flakes' $HOME/.config/nix/nix.conf

            {{- /* kl_bin_dir="/home/kl/.local/bin" */}}
            {{- /* if [ ! -f "$kl_bin_dir/kl" ]; then */}}
            {{- /*   mkdir -p $kl_bin_dir */}}
            {{- /*   pushd $kl_bin_dir */}}
            {{- /*   curl https://i.jpillora.com/kloudlite/kl@v1.1.87-nightly | bash */}}
            {{- /*   popd */}}
            {{- /* fi */}}

            mkdir -p $HOME/workspaces
            workspace_dir="$HOME/workspaces/$(KL_WORKSPACE)"
            if [ ! -d "$workspace_dir" ]; then
              mkdir -p $workspace_dir
              pushd $workspace_dir
              export PATH=$PATH:/home/kl/.nix-profile/bin:/home/kl/.local/bin 
              kl init
              popd
            fi

            if [ -f "$workspace_dir/kl.yaml" ] || [ -f "$workspace_dir/kl.yml" ]; then
              pushd $workspace_dir
              PATH=$PATH:/home/kl/.nix-profile/bin:/home/kl/.local/bin /home/kl/.local/bin/kl shell -r >  /env/.env
              PATH=$PATH:/home/kl/.nix-profile/bin:/home/kl/.local/bin /home/kl/.local/bin/kl get env >  /env/.connected_env
              popd
            fi

            if [ ! -f "/home/kl/.zshrc" ]; then
              mkdir -p "/home/kl/.config/zsh"
              cp /tmp/.zshrc /home/kl/.zshrc
              cp /tmp/.aliasrc /home/kl/.config/aliasrc
            fi

            if [ ! -f "/home/kl/.local/bin/starship" ]; then
              curl -sS https://starship.rs/install.sh | sh -s -- -y -b /home/kl/.local/bin
            fi
            
            if [ ! -d "/home/kl/.config/zsh/zsh-autosuggestions" ]; then
              mkdir -p "/home/kl/.config/zsh"
              git clone https://github.com/zsh-users/zsh-autosuggestions /home/kl/.config/zsh/zsh-autosuggestions
            fi

            if [ ! -d "/home/kl/.config/zsh/zsh-syntax-highlighting" ]; then
              mkdir -p "/home/kl/.config/zsh"
              git clone https://github.com/zsh-users/zsh-syntax-highlighting.git  "/home/kl/.config/zsh/zsh-syntax-highlighting"
            fi

            {{- /* if [ ! -d "/home/kl/.kl" ]; then */}}
            {{- /*   mkdir -p /home/kl/.kl */}}
            {{- /*   sh -c 'cat > /home/kl/.kl/kl-session.yaml <<EOF */}}
            {{- /*   session: {{.KloudliteSessionID}} */}}
            {{- /*   team: {{.KloudliteTeam}} */}}
            {{- /*   EOF' */}}
            {{- /* fi */}}
            if [ ! -f "/home/kl/.local/bin/kubectl" ]; then
              pushd /home/kl/.local/bin
              curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
              chmod +x /home/kl/.local/bin/kubectl
              popd
            fi

          volumeMounts: &volume-mounts
            - mountPath: /home/kl
              name: home-dir
            
            - mountPath: /tmp/authorized_keys
              name: ssh-keys
              subPath: authorized_keys
            
            - mountPath: /tmp/id_rsa.pub
              name: ssh-keys
              subPath: public_key
            
            - mountPath: /tmp/id_rsa
              name: ssh-keys
              subPath: private_key

            - mountPath: /nix
              name: nix-dir

            - mountPath: /env
              name: containerenv

      containers:
      {{- if .SSHContainer.Enable }}
      {{- with .SSHContainer }}
        - name: ssh
          image: {{.Image | squote}}
          imagePullPolicy: {{$.ImagePullPolicy | default "Always" }}
          env: 
            - name: KL_WORKSPACE
              value: "{{$.Metadata.Name}}"

            - name: KL_WORKSPACE_DIR
              value: "/home/kl/workspaces/{{$.Metadata.Name}}"

            - name: KL_DEVICE_NAME
              value: {{$.KloudliteDeviceFQDN}}

            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace

            - name: DEPLOYMENT_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['app']

            - name: PORT
              value: {{.Port | squote}}
          ports:
            - containerPort: {{.Port}}
          volumeMounts: *volume-mounts
      {{- end }}
      {{- end }}

      {{- if .TTYDContainer.Enable }}
      {{- with .TTYDContainer }}
        - name: ttyd
          image: {{.Image}}
          imagePullPolicy: {{$.ImagePullPolicy | default "Always"}}
          env:
            - name: KL_WORKSPACE
              value: "{{$.Metadata.Name}}"

            - name: KL_WORKSPACE_DIR
              value: "/home/kl/workspaces/{{$.Metadata.Name}}"

            - name: KL_DEVICE_NAME
              value: {{$.KloudliteDeviceFQDN}}

            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace

            - name: DEPLOYMENT_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['app']

            - name: PORT
              value: {{.Port | squote}}
          ports:
            - containerPort: {{.Port}}
          volumeMounts: *volume-mounts
      {{- end }}
      {{- end }}

      {{- if .JupyterNotebookContainer.Enable }}
      {{- with .JupyterNotebookContainer }}
        - name: jupyter-notebook
          image: {{.Image}}
          imagePullPolicy: {{$.ImagePullPolicy | default "Always"}}
          env:
            - name: KL_WORKSPACE
              value: "{{$.Metadata.Name}}"

            - name: KL_WORKSPACE_DIR
              value: "/home/kl/workspaces/{{$.Metadata.Name}}"

            - name: KL_DEVICE_NAME
              value: {{$.KloudliteDeviceFQDN}}

            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace

            - name: DEPLOYMENT_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['app']

            - name: PORT
              value: {{.Port | squote}}
          ports:
            - containerPort: {{.Port }}
          volumeMounts: *volume-mounts
          securityContext:
            runAsUser: 1000
            runAsGroup: 1000
      {{- end }}
      {{- end }}

      {{- if .VSCodeServerContainer.Enable }}
      {{- with .VSCodeServerContainer }}
        - name: vscode-server
          image: {{.Image}}
          imagePullPolicy: {{$.ImagePullPolicy | default "Always"}}
          env:
            - name: KL_WORKSPACE
              value: "{{$.Metadata.Name}}"

            - name: KL_WORKSPACE_DIR
              value: "/home/kl/workspaces/{{$.Metadata.Name}}"

            - name: KL_DEVICE_NAME
              value: {{$.KloudliteDeviceFQDN}}

            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace

            - name: DEPLOYMENT_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['app']

            - name: PORT
              value: {{.Port | squote}}
          ports:
            - containerPort: {{.Port }}
          volumeMounts: *volume-mounts
          securityContext:
            runAsUser: 1000
            runAsGroup: 1000
      {{- end }}
      {{- end }}

      {{- if .VSCodeTunnelContainer.Enable }}
      {{- with .VSCodeTunnelContainer }}
        - name: vscode-tunnel
          image: {{.Image}}
          imagePullPolicy: {{$.ImagePullPolicy | default "Always"}}
          env:
            - name: KL_WORKSPACE
              value: "{{$.Metadata.Name}}"

            - name: KL_WORKSPACE_DIR
              value: "/home/kl/workspaces/{{$.Metadata.Name}}"

            - name: KL_DEVICE_NAME
              value: {{$.KloudliteDeviceFQDN}}

            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace

            - name: DEPLOYMENT_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['app']
          volumeMounts: *volume-mounts
          securityContext:
            runAsUser: 1000
            runAsGroup: 1000
      {{- end }}
      {{- end }}

      volumes:
      - name: ssh-keys
        secret:
          secretName: {{.SSHSecretName}}
          optional: true

      - name: containerenv
        emptyDir: {}
      
      - name: home-dir
        hostPath:
          type: DirectoryOrCreate
          path: /external-volume/user-home

      - name: nix-dir
        hostPath:
          type: DirectoryOrCreate
          path: /external-volume/nix
{{- end }}
