#!/usr/bin/env sh

[ -z "$PORT" ] && echo "WARNING: PORT env-var is not defined, defaulting to 22"
PORT=${PORT:-22}

$(which sshd) -D -p "$PORT" &
pid=$!
trap 'kill -9 $pid' TERM INT EXIT
cat <<EOF
███████ ███████ ██   ██ 
██      ██      ██   ██ 
███████ ███████ ███████ 
     ██      ██ ██   ██ 
███████ ███████ ██   ██ Daemon Running on :$PORT
EOF
wait $pid
