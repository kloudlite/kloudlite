#!/usr/bin/env bash
# shellcheck source=/dev/null
set -o errexit
set -o pipefail

trap "echo kloudlite-entrypoint:CRASHED >&2" EXIT SIGINT SIGTERM

export IN_DEV_BOX="true"
export KL_WORKSPACE="$KL_WORKSPACE"
export KL_TMP_PATH="/kl-tmp"

cat <<EOL > /kl-tmp/global-profile
export SSH_PORT=$SSH_PORT
export IN_DEV_BOX="true"
export KL_WORKSPACE="$KL_WORKSPACE"
export MAIN_PATH=$PATH
export KL_TMP_PATH="/kl-tmp"
EOL

sudo dnsmasq --server=/.local/$KL_DNS --server=1.1.1.1

cat > /tmp/resolv.conf <<EOF
nameserver 127.0.0.1
search $KL_SEARCH_DOMAIN
options ndots:0
EOF

sudo cp /tmp/resolv.conf /etc/resolv.conf

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

mkdir -p ~/.config/nix
echo "experimental-features = nix-command flakes" > ~/.config/nix/nix.conf

# shift

PATH=$PATH:$HOME/.nix-profile/bin

pushd "$HOME/workspace"

cat $KL_HASH_FILE | jq '.config.env | to_entries | map_values(. = "export \(.key)=\"\(.value)\"")|.[]' -r >> /tmp/env
eval $(cat $KL_HASH_FILE | jq '.config.kloudliteConfig.mounts | to_entries | map_values(. = "printf \"\(.value)\" > \(.key)") | .[]' -r)
echo export PATH=$PATH:$(eval nix shell $(cat $KL_HASH_FILE | jq '.config.packageHashes | to_entries | map_values(. = .value) | .[]' -r | xargs -I{} printf "%s " {}) --command printenv PATH) >> /tmp/env

#source /tmp/env
#kl box reload -s && echo 'echo kloudlite-entrypoint:INSTALLING_PACKAGES_DONE'
popd

if [ -d "/tmp/ssh2" ]; then
    mkdir -p /home/kl/.ssh
    cp /tmp/ssh2/authorized_keys /home/kl/.ssh/authorized_keys
    chmod 600 /home/kl/.ssh/authorized_keys
    echo "successfully copied ssh credentials"
fi 

# sudo /mounter --conf $KL_DEVBOX_JSON_PATH

export SSH_PORT=$SSH_PORT
trap - EXIT SIGTERM SIGINT
echo "kloudlite-entrypoint: SETUP_COMPLETE"

/track-changes.sh "$KL_HASH_FILE" "echo kl-hash-file changed, exiting ...; sudo pkill -9 sshd" &

sudo /usr/sbin/sshd -D -p "$SSH_PORT"

