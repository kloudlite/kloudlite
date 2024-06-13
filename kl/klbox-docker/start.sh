#!/bin/bash
# shellcheck source=/dev/null

set -o errexit
set -o pipefail

trap "echo kloudlite-entrypoint:CRASHED >&2" EXIT SIGINT SIGTERM

export IN_DEV_BOX="true"
export KL_WORKSPACE=$KL_WORKSPACE

mkdir -p /home/kl/.kl/
cat <<EOL > /home/kl/.kl/global-profile
export SSH_PORT=$SSH_PORT
export IN_DEV_BOX="true"
export KL_WORKSPACE=$KL_WORKSPACE
export MAIN_PATH=$PATH
EOL

kl app start-dns &

cat > /tmp/resolv.conf <<'EOF'
nameserver 127.0.0.2
EOF

sudo cp /tmp/resolv.conf /etc/resolv.conf


# KL_LOCK_PATH=/home/kl/workspace/kl.lock
#
KL_DEVBOX_PATH=$HOME/.kl/devbox

mkdir -p "$KL_DEVBOX_PATH"

chown kl:kl "$HOME/.kl"

entrypoint_executed="/home/kl/.kloudlite_entrypoint_executed"
if [ ! -f "$entrypoint_executed" ]; then
    mkdir -p /home/kl/.config
    cp /tmp/.zshrc /home/kl/.zshrc
    cp /tmp/.bashrc /home/kl/.bashrc
    cp /tmp/.profile /home/kl/.profile
    ln -sf /home/kl/.profile /home/kl/.zprofile
    cp /tmp/aliasrc /home/kl/.config/aliasrc
    echo "successfully initialized .profile and .bashrc" >> $entrypoint_executed
fi

shift

PATH=$PATH:$HOME/.nix-profile/bin

pushd "$HOME/workspace"
echo "kloudlite-entrypoint:INSTALLING_PACKAGES"
kl box reload -s && echo 'echo kloudlite-entrypoint:INSTALLING_PACKAGES_DONE'
popd

if [ -d "/tmp/ssh2" ]; then
    mkdir -p /home/kl/.ssh
    cp /tmp/ssh2/authorized_keys /home/kl/.ssh/authorized_keys
    chmod 600 /home/kl/.ssh/authorized_keys
    echo "successfully copied ssh credentials"
fi 

# sudo /mounter --conf $KL_DEVBOX_JSON_PATH

export SSH_PORT=$SSH_PORT
# trap - EXIT SIGTERM SIGINT
echo "kloudlite-entrypoint: SETUP_COMPLETE"
sudo /usr/sbin/sshd -D -p "$SSH_PORT" 
