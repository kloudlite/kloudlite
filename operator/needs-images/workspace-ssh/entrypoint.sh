#!/usr/bin/env sh

$(which sshd) -D &
pid=$!
trap 'kill -9 $pid' TERM INT EXIT
cat <<EOF
███████ ███████ ██   ██ 
██      ██      ██   ██ 
███████ ███████ ███████ 
     ██      ██ ██   ██ 
███████ ███████ ██   ██ Daemon Running
EOF
wait $pid
