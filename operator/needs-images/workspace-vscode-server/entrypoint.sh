#!/usr/bin/env bash

[ -z "$KL_WORKSPACE_DIR" ] && echo "env var: KL_WORKSPACE_DIR is missing" && exit 1
[ -z "$KL_WORKSPACE" ] && echo "env var: KL_WORKSPACE is missing" && exit 1

[ ! -d "$KL_WORKSPACE_DIR" ] && echo "workspace dir: $KL_WORKSPACE_DIR does not exist" && exit 1

cd "$KL_WORKSPACE_DIR"

# sudo chown -R kl /home/kl

code serve-web --host "0.0.0.0" --cli-data-dir "$HOME/.vscode-web/$KL_WORKSPACE" --accept-server-license-terms --without-connection-token &

pid=$!
trap 'kill -9 $pid' TERM INT EXIT
cat <<EOF
▖▖▄▖▄▖   ▌    ▄▖          
▌▌▚ ▌ ▛▌▛▌█▌  ▚ █▌▛▘▌▌█▌▛▘
▚▘▄▌▙▖▙▌▙▌▙▖  ▄▌▙▖▌ ▚▘▙▖▌  
EOF
wait $pid
