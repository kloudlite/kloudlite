#! /usr/bin/env bash

cat > ~/runner-config.yml <<EOF2
runAs: agent
agent:
  # publicIP: ${tf_public_ip}
  serverIP: ${tf_k3s_masters_dns_host}
  token: ${tf_k3s_token}
  labels: ${tf_node_labels}
  nodeName: ${tf_node_name}
  taints: ${jsonencode([
    for taint in  tf_node_taints: "${taint.key}=${taint.value}:${taint.effect}"
  ])}
  extraAgentArgs: ${jsonencode(tf_extra_agent_args)}
EOF2

sudo ln -sf ~/runner-config.yml /runner-config.yml
if [ "${tf_use_cloudflare_nameserver}" = "true" ]; then
lineNo=$(sudo cat /etc/resolv.conf -n | grep "nameserver" | awk '{print $1}')
sudo sed -i "$lineNo i nameserver 1.1.1.1" /etc/resolv.conf
fi
