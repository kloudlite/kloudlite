#!/usr/bin/env bash
# shellcheck source=/dev/null
set -o errexit
set -o pipefail

trap "echo kloudlite-entrypoint:CRASHED >&2" EXIT SIGINT SIGTERM

#/nix-installer/install

export IN_DEV_BOX="true"
export KL_WORKSPACE="$KL_WORKSPACE"
export KL_TMP_PATH="/kl-tmp"

cat <<EOL > /kl-tmp/global-profile
export SSH_PORT=$SSH_PORT
export IN_DEV_BOX="true"
export KL_WORKSPACE="$KL_WORKSPACE"
export KL_BASE_URL="$KL_BASE_URL"
export MAIN_PATH=$PATH
export KL_TMP_PATH="/kl-tmp"
export KLCONFIG_PATH="$KLCONFIG_PATH"
EOL

sudo dnsmasq --server=/.local/$KL_DNS --server=1.1.1.1

sudo chown kl /var/run/docker.sock

entrypoint_executed="/home/kl/.kloudlite_entrypoint_executed"
if [ ! -f "$entrypoint_executed" ]; then
    mkdir -p /home/kl/.config
    cp /tmp/.zshrc /home/kl/.zshrc
    cp /tmp/.bashrc /home/kl/.bashrc
    cp /tmp/.profile /home/kl/.profile
    ln -sf /home/kl/.profile /home/kl/.zprofile
    cp /tmp/aliasrc /home/kl/.config/aliasrc
    echo "successfully initialized .profile and .bashrc" >> $entrypoint_executed
    ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa -N "" <<<y >/dev/null 2>&1
fi

mkdir -p ~/.config/nix
echo "experimental-features = nix-command flakes" > ~/.config/nix/nix.conf

# shift

PATH=$PATH:$HOME/.nix-profile/bin

echo "kloudlite-entrypoint:INSTALLING_PACKAGES"
cat $KL_HASH_FILE | jq '.hash' -r > /tmp/hash
cat $KL_HASH_FILE | jq '.config.env | to_entries | map_values(. = "export \(.key)=\"\(.value)\"")|.[]' -r >> /tmp/env
cat > /tmp/mount.sh <<EOF
vmounts=$(cat $KL_HASH_FILE | jq '.config.mounts | length')
if [ \$vmounts -gt 0 ]; then
  eval $(cat $KL_HASH_FILE | jq '.config.mounts | to_entries | map_values(. = "mkdir -p $(dirname \(.key))") | .[]' -r)
  eval $(cat $KL_HASH_FILE | jq '.config.mounts | to_entries | map_values(. = "echo \"\(.value)\" > \(.key)") | .[]' -r)
fi
EOF
sudo bash /tmp/mount.sh

cat > /tmp/pkg-install.sh <<EOF
npkgs=$(cat $KL_HASH_FILE | jq '.config.packageHashes | length')
if [ \$npkgs -gt 0 ]; then
  nix shell --log-format bar-with-logs $(cat $KL_HASH_FILE | jq '.config.packageHashes | to_entries | map_values(. = .value) | .[]' -r | xargs -I{} printf "%s " {}) --command echo "successfully installed packages"
  npath=$(nix shell $(cat $KL_HASH_FILE | jq '.config.packageHashes | to_entries | map_values(. = .value) | .[]' -r | xargs -I{} printf "%s " {}) --command printenv PATH)
  echo export PATH=$PATH:\$npath >> /tmp/env
fi
EOF

bash /tmp/pkg-install.sh

#nix shell $(cat $KL_HASH_FILE | jq '.config.packageHashes | to_entries | map_values(. = .value) | .[]' -r | xargs -I{} printf "%s " {}) --command echo "successfully installed packages"
#echo export PATH=$PATH:$(eval nix shell $(cat $KL_HASH_FILE | jq '.config.packageHashes | to_entries | map_values(. = .value) | .[]' -r | xargs -I{} printf "%s " {}) --command printenv PATH) >> /tmp/env
echo "export KL_HASH_FILE=$KL_HASH_FILE" >> /tmp/env
echo "kloudlite-entrypoint:INSTALLING_PACKAGES_DONE"

source /tmp/env

cat > /tmp/resolv.conf <<EOF
nameserver 127.0.0.1
search $KL_SEARCH_DOMAIN
options ndots:0
EOF

sudo cp /tmp/resolv.conf /etc/resolv.conf


if [ -d "/tmp/ssh2" ]; then
    mkdir -p /home/kl/.ssh
    cp /tmp/ssh2/authorized_keys /home/kl/.ssh/authorized_keys
    chmod 600 /home/kl/.ssh/authorized_keys
    echo "successfully copied ssh credentials"
fi 

export SSH_PORT=$SSH_PORT
trap - EXIT SIGTERM SIGINT
echo "kloudlite-entrypoint:SETUP_COMPLETE"

sudo /usr/sbin/sshd -D -p "$SSH_PORT"
