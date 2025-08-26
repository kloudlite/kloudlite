#!/usr/bin/env bash
set -e
if [ -f "/env/.connected_env" ]; then
  cat >/tmp/resolv.conf <<EOF
search env-$(cat /env/.connected_env).svc.cluster.local svc.cluster.local cluster.local
nameserver 10.43.0.10
options ndots:5
EOF
  sudo cp /tmp/resolv.conf /etc/resolv.conf
fi

export HOST=box
export KL_BOX_MODE=true
if [ -f "/env/.env" ]; then
  export $(grep -v '^#' /env/.env | xargs)
  sudo bash -c "cat /env/.env >> /etc/environment"
fi

[ -z "$KL_WORKSPACE_DIR" ] && echo "env var: KL_WORKSPACE_DIR is missing"

cd "$KL_WORKSPACE_DIR"

CODE_SERVER_PORT=8080

code-server --bind-addr "0.0.0.0:$CODE_SERVER_PORT" --auth none &
pid=$!
trap 'kill -9 $pid' TERM INT EXIT
wait $pid
