{{- with . }}
replicas: 1
selector:
  matchLabels: {{ .SelectorLabels | toJson }}
template:
  metadata:
    labels: {{.SelectorLabels | toJson }}
  spec:
    containers:
      - name: sshd-server
        image: ghcr.io/kloudlite/hub/ssh-server
        ports:
          - containerPort: 22  # Internal port used by the image
        volumeMounts:
          - name: ssh-secret
            mountPath: /home/kl/.ssh/authorized_keys
            subPath: authorized_keys

    volumes:
      - name: ssh-secret
        secret:
          secretName: {{.SSHAuthorizedKeysSecretName}}
          items:
            - key: "{{.SSHAuthorizedKeysSecretKey}}"
              path: "authorized_keys"
{{- end }}
