#!/usr/bin/env bash

# set -e
# if [ -f "/env/.connected_env" ]; then
#   cat >/tmp/resolv.conf <<EOF
# search env-$(cat /env/.connected_env).svc.cluster.local svc.cluster.local cluster.local
# nameserver 10.43.0.10
# options ndots:5
# EOF
#   sudo cp /tmp/resolv.conf /etc/resolv.conf
# fi

# export HOST=box
# export KL_BOX_MODE=true
# if [ -f "/env/.env" ]; then
#   export $(grep -v '^#' /env/.env | xargs)
#   sudo bash -c "cat /env/.env >> /etc/environment"
# fi

[ -z "$KL_WORKSPACE_DIR" ] && echo "env var: KL_WORKSPACE_DIR is missing" && exit 1
[ -z "$KL_WORKSPACE" ] && echo "env var: KL_WORKSPACE is missing" && exit 1

[ ! -d "$KL_WORKSPACE_DIR" ] && echo "workspace dir: $KL_WORKSPACE_DIR does not exist" && exit 1

cd "$KL_WORKSPACE_DIR"

# sudo chown -R kl /home/kl

mkdir -p "$HOME/.tunnels/$KL_WORKSPACE"
/vs-code/code tunnel --name "$KL_WORKSPACE" --cli-data-dir "$HOME/.tunnels/$KL_WORKSPACE" --accept-server-license-terms &

pid=$!
trap 'kill -9 $pid' TERM INT EXIT
cat <<EOF
▖▖▄▖▄▖   ▌    ▄▖          
▌▌▚ ▌ ▛▌▛▌█▌  ▚ █▌▛▘▌▌█▌▛▘
▚▘▄▌▙▖▙▌▙▌▙▖  ▄▌▙▖▌ ▚▘▙▖▌  
EOF
wait $pid
