#!/bin/bash

set -o errexit
set -o pipefail

trap "echo kloudlite-entrypoint:CRASHED >&2" EXIT SIGINT SIGTERM

KL_DEVBOX_PATH=/home/kl/.kl/devbox
mkdir -p $KL_DEVBOX_PATH


if [ -f "/home/kl/workspace/kl.lock" ]; then
    cp /home/kl/workspace/kl.lock $KL_DEVBOX_PATH/devbox.lock
fi


entrypoint_executed="/home/kl/.kloudlite_entrypoint_executed"
if [ ! -f "$entrypoint_executed" ]; then
    cp /tmp/.bashrc /home/kl/.bashrc
    cp /tmp/.profile /home/kl/.profile
    echo "successfully initialized .profile and .bashrc" >> $entrypoint_executed
fi

shift
echo "$@" | jq -r > /tmp/sample.json

PATH=$PATH:$HOME/.nix-profile/bin
mkdir -p $KL_DEVBOX_PATH
cp /tmp/sample.json $KL_DEVBOX_PATH/devbox.json
cd $KL_DEVBOX_PATH

echo "kloudlite-entrypoint:INSTALLING_PACKAGES"
devbox install && devbox update && echo 'echo kloudlite-entrypoint:INSTALLING_PACKAGES_DONE'

if [ -d "/tmp/ssh2" ]; then
    mkdir -p /home/kl/.ssh
    cp /tmp/ssh2/authorized_keys /home/kl/.ssh/authorized_keys
    chmod 600 /home/kl/.ssh/authorized_keys
    echo "successfully copied ssh credentials"
fi 

sudo /mounter --conf /tmp/sample.json

mkdir -p /home/kl/.kl
cat <<EOL > /home/kl/.kl/global-profile
export SSH_PORT=$SSH_PORT
export IN_DEV_BOX="true"
EOL

# trap - EXIT SIGTERM SIGINT
echo "kloudlite-entrypoint: SETUP_COMPLETE"
sudo /usr/sbin/sshd -D -p $SSH_PORT
