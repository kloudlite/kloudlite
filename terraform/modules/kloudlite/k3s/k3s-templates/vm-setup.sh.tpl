#! /usr/bin/env bash

rm -rf ${kloudlite_config_directory}/execution.log
debug() {
  echo "[#] $*" >> ${kloudlite_config_directory}/execution.log
}

debug "execution started at $(date)"

[ $EUID -ne 0 ] && echo "this script must be run as root. current EUID is $EUID" && exit 1

arch=$(uname -m)
if [ "$arch" != "x86_64" ]; then
	echo "CPU Architecture is not x86_64. Exiting ..." && exit 1
fi

debug "ensuring ${kloudlite_config_directory} exists"
mkdir -p ${kloudlite_config_directory}

debug "downloading from kloudite release ${kloudlite_release}: k3s binary"
curl -L0 "https://github.com/kloudlite/infrastructure-as-code/releases/download/${kloudlite_release}/k3s" >/usr/local/bin/k3s
chmod +x /usr/local/bin/k3s

debug "downloading from kloudite release ${kloudlite_release}: kloudlite runner binary"
curl -L0 "https://github.com/kloudlite/infrastructure-as-code/releases/download/${kloudlite_release}/runner-amd64" >/usr/local/bin/kloudlite-runner
chmod +x /usr/local/bin/kloudlite-runner

debug "creating SystemD service: kloudlite-k3s.service /etc/systemd/system/kloudlite-k3s.service"
cat >/etc/systemd/system/kloudlite-k3s.service <<EOF
[Unit]
Description=This script will start kloudlite k3s runner. It is maintained by kloudlite.io, and is used to run k3s with a custom set of args.

[Service]
ExecStart=kloudlite-runner --config ${kloudlite_config_directory}/runner-config.yml

[Install]
WantedBy=multi-user.target
EOF

systemctl enable --now kloudlite-k3s.service