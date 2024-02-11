#! /usr/bin/env bash

sleep 30s

curl -L0 "https://github.com/k3s-io/k3s/releases/download/${k3s_release}/k3s" >/usr/local/bin/k3s
chmod +x /usr/local/bin/k3s

curl -L0 "https://github.com/kloudlite/infrastructure-as-code/releases/download/${kloudlite_release}/runner-amd64" >./runner
chmod +x ./runner

cat >/etc/systemd/system/kloudlite-k3s.service <<EOF
[Unit]
Description=This script will start kloudlite k3s runner. It is maintaind by kloudlite.io, and is used to run k3s with a custom set of args.

[Service]
ExecStart=$PWD/runner --config $PWD/runner-config.yml

[Install]
WantedBy=multi-user.target
EOF
