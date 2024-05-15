#! /usr/bin/env bash

${vm_setup_script}

cat > ${kloudlite_config_directory}/runner-config.yml <<EOF2
${jsonencode({
  "k3s_flags": concat(
    [
      "agent",
      "--server", "https://${tf_k3s_masters_dns_host}:6443",
      "--token", "${tf_k3s_token}",
      "--node-name", "${tf_node_name}",
    ],
    flatten([ for k,v in tf_node_labels: ["--node-label", "${k}=${v}"] ]),
    flatten([ for taint in tf_node_taints: ["--node-taint", "${taint.key}=${taint.value}:${taint.effect}"] ]),
    ["--kubelet-arg", "--system-reserved=cpu=100m,memory=200Mi,ephemeral-storage=1Gi,pid=1000"],
    tf_extra_agent_args,
  ),
})}
EOF2

if [ "${tf_use_cloudflare_nameserver}" = "true" ]; then
lineNo=$(sudo cat /etc/resolv.conf -n | grep "nameserver" | awk '{print $1}')
sudo sed -i "$lineNo i nameserver 1.1.1.1" /etc/resolv.conf
fi

sudo systemctl restart kloudlite-k3s.service
