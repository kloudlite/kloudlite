#! /usr/bin/env bash

set -o errexit
set -o pipefail

/docker-socket.sh &

usermod -u $HOST_USER_UID kl
#usermod -g $HOST_USER_GID kl
chown -R kl:kl /kl-tmp
chown -R kl:kl /nix

# sleep 1000000

/start.sh

export SSH_PORT=$SSH_PORT
/usr/sbin/sshd -D -p "$SSH_PORT"
