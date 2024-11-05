#!/usr/bin/env bash

#sudo echo "${node_name}" > /etc/hostname

cat > ~/runner-config.yml <<EOF
runAs: agent
agent:
  serverIP: ${k3s_server_host}
  token: ${k3s_token}
  labels: ${jsonencode(node_labels)}
  nodeName: ${node_name}
EOF

sudo ln -sf ~/runner-config.yml /runner-config.yml
if [ "${disable_ssh}" == "true" ]; then
  sudo systemctl disable sshd.service
  sudo systemctl stop sshd.service
  sudo rm -f ~/.ssh/authorized_keys
fi