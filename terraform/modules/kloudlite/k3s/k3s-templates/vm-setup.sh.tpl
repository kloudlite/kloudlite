#! /usr/bin/env bash

KLOUDLITE_CONFIG_DIRECTORY=${kloudlite_config_directory}
KLOUDLITE_RELEASE=${kloudlite_release}

LOG_FILE=$KLOUDLITE_CONFIG_DIRECTORY/execution.log

debug() {
  echo "[#] $*" >> $LOG_FILE
}

BIN_DIR=/usr/local/bin

mkdir -p $KLOUDLITE_CONFIG_DIRECTORY
debug "----------------- execution started at $(date) ----------------------"
[ $EUID -ne 0 ] && debug "this script must be run as root. current EUID is $EUID" && exit 1

arch=$(uname -m)
if [ "$arch" != "x86_64" ]; then
	echo "CPU Architecture is not x86_64. Exiting ..." && exit 1
fi

debug "ensuring $KLOUDLITE_CONFIG_DIRECTORY exists"

cat > $BIN_DIR/kloudlite-upgrade.sh <<'EOF'
#! /usr/bin/env bash

k3s_bin_path=/usr/local/bin/k3s
kloudlite_runner_bin_path=/usr/local/bin/kloudlite-runner
KLOUDLITE_RELEASE=$1

systemctl is-active --quiet kloudlite-k3s.service && systemctl stop kloudlite-k3s.service && (pkill kloudlite-runner; pkill k3s)

#BASE_URL="https://github.com/kloudlite/infrastructure-as-code/releases/download/$KLOUDLITE_RELEASE"
BASE_URL="https://github.com/kloudlite/kloudlite/releases/download/$KLOUDLITE_RELEASE"

echo "[#] downloading from kloudlite release $KLOUDLITE_RELEASE: k3s binary"
curl -L0 "$BASE_URL/k3s" >$k3s_bin_path
chmod +x $k3s_bin_path
echo "[#] downloaded @ $k3s_k3s_bin_path"

echo "[#] downloading from kloudlite release $KLOUDLITE_RELEASE: kloudlite runner binary"
curl -L0 "$BASE_URL/runner-amd64" >$kloudlite_runner_bin_path
chmod +x $kloudlite_runner_bin_path
debug "[#] downloaded @ $kloudlite_runner_bin_path"

EOF

chmod +x $BIN_DIR/kloudlite-upgrade.sh >> $LOG_FILE

kloudlite-upgrade.sh $KLOUDLITE_RELEASE

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
debug "----------------- execution finished at $(date) ----------------------"
