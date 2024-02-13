#! /usr/bin/env bash

[ $EUID -ne 0 ] && echo "this script must be run as root. current EUID is $EUID"

arch=$(uname -m)
if [ "$arch" != "x86_64" ]; then
	echo "CPU Architecture is not x86_64. Exiting ..." && exit 1
fi

#sleep 30s

kloudlite_release=${KLOUDLITE_RELEASE}

#curl -L0 "https://github.com/k3s-io/k3s/releases/download/$k3s_release/k3s" >/usr/local/bin/k3s
#chmod +x /usr/local/bin/k3s

curl -L0 "https://github.com/kloudlite/infrastructure-as-code/releases/download/$kloudlite_release/k3s" >./usr/local/bin/k3s
chmod +x /usr/local/bin/k3s

curl -L0 "https://github.com/kloudlite/infrastructure-as-code/releases/download/$kloudlite_release/runner-amd64" >./runner
chmod +x ./runner

cat >/etc/systemd/system/kloudlite-k3s.service <<EOF
[Unit]
Description=This script will start kloudlite k3s runner. It is maintained by kloudlite.io, and is used to run k3s with a custom set of args.

[Service]
ExecStart=$PWD/runner --config $PWD/runner-config.yml

[Install]
WantedBy=multi-user.target
EOF

systemctl enable --now kloudlite-k3s.service