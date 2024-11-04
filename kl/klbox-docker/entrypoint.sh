#! /usr/bin/env bash

set -o errexit
set -o pipefail

/docker-socket.sh &

/start.sh

export SSH_PORT=$SSH_PORT
/usr/sbin/sshd -D -p "$SSH_PORT"
