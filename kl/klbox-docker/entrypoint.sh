#! /usr/bin/env bash

set -o errexit
set -o pipefail

sudo /docker-socket.sh &

/start.sh

export SSH_PORT=$SSH_PORT
sudo /usr/sbin/sshd -D -p "$SSH_PORT"
