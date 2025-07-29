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

            if [ ! -d "/home/kl/.ssh" ]; then
              mkdir -p /home/kl/.ssh
              chown -R 1000:1000 /home/kl/.ssh
            fi


            if [ -f "/home/kl/.ssh/authorized_keys" ]; then
              if ! cmp -s /tmp/authorized_keys /home/kl/.ssh/authorized_keys; then
                echo "authorized_keys file differs, copying new one"
                cp /tmp/authorized_keys /home/kl/.ssh/authorized_keys
              fi
              echo "authorized_keys file is up to date"
            else
              echo "authorized_keys file not found, copying new one"
              {{- /* rm -rf /home/kl/.ssh/authorized_keys */}}
              cp /tmp/authorized_keys /home/kl/.ssh/authorized_keys
            fi
            
            if [ -f "/home/kl/.ssh/id_rsa" ]; then
              if ! cmp -s /tmp/id_rsa /home/kl/.ssh/id_rsa; then
                echo "id_rsa file differs, copying new one"
                rm -f /home/kl/.ssh/id_rsa* 2>/dev/null || true
                cp /tmp/id_rsa /home/kl/.ssh/id_rsa
                cp /tmp/id_rsa.pub /home/kl/.ssh/id_rsa.pub
              fi
              echo "id_rsa file is up to date"
            else
              echo "id_rsa file not found, copying new one"
              rm -f /home/kl/.ssh/id_rsa* 2>/dev/null || true
              cp /tmp/id_rsa /home/kl/.ssh/id_rsa
              cp /tmp/id_rsa.pub /home/kl/.ssh/id_rsa.pub
            fi
            
            if [ ! -d "/nix/store" ]; then
              curl -L https://nixos.org/nix/install | sh
              mkdir -p ~/.config/nix
              echo 'experimental-features = nix-command flakes' > ~/.config/nix/nix.conf
            fi
            kl_bin_dir="/home/kl/.local/bin"
            if [ ! -f "$kl_bin_dir/kl" ]; then
              mkdir -p $kl_bin_dir
              pushd $kl_bin_dir
              curl https://i.jpillora.com/kloudlite/kl@v1.1.87-nightly | bash
              popd
            fi

            workspace_dir="/home/kl/workspaces/$(KL_WORKSPACE)"
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
        - name: ssh
          image: {{.ImageSSH | squote}}
          imagePullPolicy: {{.ImagePullPolicy | default "IfNotPresent" }}
          env: &env
            - name: KL_WORKSPACE
              value: "{{.Metadata.Name}}"

            - name: KL_WORKSPACE_DIR
              value: "/home/kl/workspaces/{{.Metadata.Name}}"

            - name: KL_DEVICE_NAME
              value: {{.KloudliteDeviceFQDN}}

            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace

            - name: DEPLOYMENT_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.labels['app']
          ports:
            - containerPort: 22
          volumeMounts: *volume-mounts

      {{ if .EnableTTYD }}
        - name: ttyd
          image: {{.ImageTTYD}}
          imagePullPolicy: {{.ImagePullPolicy}}
          env: *env
          ports:
          - containerPort: 54535
          volumeMounts: *volume-mounts
      {{ end }}

      {{ if .EnableJupyterNotebook }}
        - name: jupyter
          image: {{.ImageJupyterNotebook}}
          imagePullPolicy: {{.ImagePullPolicy}}
          env: *env
          ports:
          - containerPort: 8888
          volumeMounts: *volume-mounts
          securityContext:
            runAsUser: 1000
            runAsGroup: 1000
      {{ end }}

      {{ if .EnableCodeServer }}
        - name: code-server
          image: {{.ImageCodeServer}}
          imagePullPolicy: {{.ImagePullPolicy}}
          env: *env
          volumeMounts: *volume-mounts
          securityContext:
            runAsUser: 1000
            runAsGroup: 1000
      {{ end }}

      {{ if .EnableVSCodeServer }}
        - name: vscode-server
          {{- /* image: ghcr.io/kloudlite/iac/vscode-server:latest */}}
          image: {{.ImageVSCodeServer}}
          imagePullPolicy: {{.ImagePullPolicy}}
          env: *env
          volumeMounts: *volume-mounts
          securityContext:
            runAsUser: 1000
            runAsGroup: 1000
      {{ end }}

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
