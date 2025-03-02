#! /usr/bin/env bash

KLOUDLITE_CONFIG_DIRECTORY=/etc/kloudlite

K3S_SERVER_HOST="${k3s_server_host}"

K3S_SERVER_TOKEN="${k3s_server_token}"
K3S_AGENT_TOKEN="${k3s_agent_token}"
# --tf params:END

debug() {
  echo "[#] $*" | tee -a "$KLOUDLITE_CONFIG_DIRECTORY/execution.log"
}

debug "ensure kloudlite config directory ($KLOUDLITE_CONFIG_DIRECTORY) exists"
mkdir -p "$KLOUDLITE_CONFIG_DIRECTORY"

debug "################# execution started at $(date) ######################"
[ $EUID -ne 0 ] && debug "this script must be run as root. current EUID is $EUID" && exit 1

create_k3s_config_file() {
  cat >"$KLOUDLITE_CONFIG_DIRECTORY/k3s.yaml" <<EOF
cluster-init: true
server: "https://$K3S_SERVER_HOST:6443"
token: "$K3S_SERVER_TOKEN"
agent-token: "$K3S_AGENT_TOKEN"

bind-address: 10.17.17.1
advertise-address: 10.17.17.1
node-ip: 10.17.17.1

tls-san-security: true
tls-san:
  - $K3S_SERVER_HOST

flannel-iface: kubernetes
flannel-backend: "wireguard-native"
write-kubeconfig-mode: "0644"
node-label:
  - "kloudlite.io/node.ip=$K3S_SERVER_HOST"

etcd-snapshot-compress: true
etcd-snapshot-schedule-cron: "1 2/2 * * *"

disable-helm-controller: true

disable: 
  - "traefik"

kubelet-arg: "--system-reserved=cpu=250m,memory=250Mi,ephemeral-storage=5Gi,pid=1000"
EOF

  mkdir -p /etc/rancher/k3s
  ln -sf $KLOUDLITE_CONFIG_DIRECTORY/k3s.yaml /etc/rancher/k3s/config.yaml
}

install_k3s() {
  debug "installing k3s latest stable version"
  curl -sfL https://get.k3s.io | INSTALL_K3S_CHANNEL=stable sh -
}

new_network_interface() {
  sudo ip link add kubernetes type veth peer name kubernetes
  sudo ip addr add 10.17.17.0/24 dev kubernetes
  sudo ip link set kubernetes up

  sudo apt update
  sudo apt install -y ufw
  sudo ufw enable
  echo "1" >/proc/sys/net/ipv4/ip_forward

  sudo ufw default reject incoming
  sudo ufw allow 80/tcp
  sudo ufw allow 22
}

create_k3s_config_file
install_k3s

debug "################# execution finished at $(date) ######################"
