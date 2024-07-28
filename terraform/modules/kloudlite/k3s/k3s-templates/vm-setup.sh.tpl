#! /usr/bin/env bash

# --tf params:START
KLOUDLITE_CONFIG_DIRECTORY=${kloudlite_config_directory}

K3S_DOWNLOAD_URL=${k3s_download_url}
KLOUDLITE_RUNNER_DOWNLOAD_URL="${kloudlite_runner_download_url}"
# --tf params:END

# LOG_FILE=$KLOUDLITE_CONFIG_DIRECTORY/execution.log

debug() {
  echo "[#] $*" >>"$KLOUDLITE_CONFIG_DIRECTORY/execution.log"
}

debug "ensuring $KLOUDLITE_CONFIG_DIRECTORY exists"
mkdir -p "$KLOUDLITE_CONFIG_DIRECTORY"

BIN_DIR=/usr/local/bin
SYSTEMD_SERVICE_PATH="/etc/systemd/system/kloudlite-k3s.service"

debug "----------------- execution started at $(date) ----------------------"
[ $EUID -ne 0 ] && debug "this script must be run as root. current EUID is $EUID" && exit 1

# arch=$(uname -m)
# if [ "$arch" != "x86_64" ]; then
# 	echo "CPU Architecture is not x86_64. Exiting ..." && exit 1
# fi

BIN_DIR=$BIN_DIR

cat >$BIN_DIR/kloudlite-install-or-upgrade.sh <<EOF
#! /usr/bin/env bash

systemctl is-active --quiet kloudlite-k3s.service && systemctl stop kloudlite-k3s.service && (pkill kloudlite-runner; pkill k3s)

echo "[#] downloading k3s binary from '$K3S_DOWNLOAD_URL'"
curl -L0 "$K3S_DOWNLOAD_URL" >$BIN_DIR/k3s
chmod +x $BIN_DIR/k3s
echo "[#] downloaded @ $BIN_DIR/k3s"

echo "[#] downloading from kloudlite release $KLOUDLITE_RELEASE: kloudlite runner binary"
curl -L0 "$KLOUDLITE_RUNNER_DOWNLOAD_URL" >$BIN_DIR/kloudlite-runner
chmod +x $BIN_DIR/kloudlite-runner
echo "[#] downloaded @ $BIN_DIR/kloudlite-runner"
EOF

chmod +x $BIN_DIR/kloudlite-install-or-upgrade.sh
kloudlite-install-or-upgrade.sh

debug "creating SystemD service: kloudlite-k3s.service /etc/systemd/system/kloudlite-k3s.service"
cat >/etc/systemd/system/kloudlite-k3s.service <<EOF
[Unit]
Description=This script will start kloudlite k3s runner. It is maintained by kloudlite.io, and is used to run k3s with a custom set of args.

[Service]
ExecStart=kloudlite-runner --config $KLOUDLITE_CONFIG_DIRECTORY/runner-config.yml

[Install]
WantedBy=multi-user.target
EOF

systemctl enable --now kloudlite-k3s.service

systemctl stop systemd-resolved
systemctl disable systemd-resolved

rm /etc/resolv.conf
echo "nameserver 1.1.1.1" >/etc/resolv.conf
debug "----------------- execution finished at $(date) ----------------------"
