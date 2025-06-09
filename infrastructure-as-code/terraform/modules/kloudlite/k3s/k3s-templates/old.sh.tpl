#! /usr/bin/env bash

# KLOUDLITE_CONFIG_DIRECTORY=${kloudlite_config_directory}
KLOUDLITE_CONFIG_DIRECTORY=/etc/kloudlite

K3S_DOWNLOAD_URL=${k3s_download_url}
# KLOUDLITE_RUNNER_DOWNLOAD_URL="${kloudlite_runner_download_url}"

K3S_SERVER_HOST="${k3s_server_host}"
K3S_SERVER_INTERNAL_IP="${k3s_server_internal_ip}"

K3S_SERVER_TOKEN="${k3s_server_token}"
K3S_AGENT_TOKEN="${k3s_agent_token}"
# --tf params:END

debug() {
  echo "[#] $*" | tee -a "$KLOUDLITE_CONFIG_DIRECTORY/execution.log"
}

debug "ensure kloudlite config directory ($KLOUDLITE_CONFIG_DIRECTORY) exists"
mkdir -p "$KLOUDLITE_CONFIG_DIRECTORY"

BIN_DIR=/usr/local/bin
SYSTEMD_SERVICE_PATH="/etc/systemd/system/kloudlite-k3s.service"

debug "################# execution started at $(date) ######################"
[ $EUID -ne 0 ] && debug "this script must be run as root. current EUID is $EUID" && exit 1

install_k3s_binary() {
  debug "downloading k3s binary from url ($K3S_DOWNLOAD_URL)"
  curl -L0 "$K3S_DOWNLOAD_URL" >$BIN_DIR/k3s
  chmod +x $BIN_DIR/k3s
  debug "[#] downloaded k3s binary @ $BIN_DIR/k3s"
}

install_kloudlite_k3s_runner() {
  debug "downloading kloudlite k3s runner from url ($KLOUDLITE_RUNNER_DOWNLOAD_URL)"
  curl -L0 "$KLOUDLITE_RUNNER_DOWNLOAD_URL" >$BIN_DIR/kloudlite-k3s-runner
  chmod +x $BIN_DIR/kloudlite-k3s-runner
  echo "[#] downloaded kloudlite k3s runner @ $BIN_DIR/kloudlite-k3s-runner"
}

creating_systemd_service() {
  debug "creating SystemD service ($SYSTEMD_SERVICE_PATH)"
  cat >$SYSTEMD_SERVICE_PATH <<EOF
[Unit]
Description=This script will start kloudlite k3s runner. It is maintained by kloudlite.io, and is used to run k3s with a custom set of args.

[Service]
ExecStart=kloudlite-k3s-runner --config $KLOUDLITE_CONFIG_DIRECTORY/runner-config.yml

[Install]
WantedBy=multi-user.target
EOF

  systemctl enable --now kloudlite-k3s.service

  systemctl stop systemd-resolved
  systemctl disable systemd-resolved
}

create_k3s_config_file() {
  echo "${K3S_SERVER_INTERNAL_IP} ${K3S_SERVER_HOST}" >>/etc/hosts

  cat >"$KLOUDLITE_CONFIG_DIRECTORY/k3s.yaml" <<EOF
cluster-init: true
server: "https://${K3S_SERVER_HOST}:6443"
token: "${K3S_SERVER_TOKEN}"
agent-token: "${K3S_AGENT_TOKEN}"

tls-san-security: true
tls-san:
  - ${K3S_SERVER_HOST}

flannel-backend: "wireguard-native"
write-kubeconfig-mode: "0644"
node-label:
  - "kloudlite.io/node.ip=${K3S_SERVER_HOST}"

etcd-snapshot-compress: true
etcd-snapshot-schedule-cron: "1 2/2 * * *"

disable-helm-controller: true

disable: 
  - "traefik"

kubelet-arg: "--system-reserved=cpu=250m,memory=250Mi,ephemeral-storage=5Gi,pid=1000"
EOF
}

# install_k3s_binary
# install_kloudlite_k3s_runner

creating_systemd_service

debug "################# execution finished at $(date) ######################"
