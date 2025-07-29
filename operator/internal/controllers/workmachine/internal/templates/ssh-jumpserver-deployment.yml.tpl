apiVersion: apps/v1
kind: Deployment
metadata: {{.Metadata | toJson}}
spec:
  replicas: 1
  selector:
    matchLabels: {{ .SelectorLabels | toJson }}
  template:
    metadata:
      labels: {{.SelectorLabels | toJson }}
    spec:
      nodeName: {{.WorkMachineName}}

      tolerations:
        - key: {{.WorkMachineTolerationKey}}
          operator: "Equal"
          value: {{.WorkMachineName |squote}}
          effect: "NoExecute"

      containers:
        - name: sshd-server
          image: {{.ImageSSHServer}}
          ports:
            - containerPort: 22  # Internal port used by the image
          volumeMounts:
            - name: ssh-secret
              mountPath: /home/kl/.ssh/authorized_keys
              subPath: authorized_keys

      volumes:
        - name: ssh-secret
          secret:
            secretName: {{.SSH.Secret.Name}}
            items:
              - key: "authorized_keys"
                path: "authorized_keys"
