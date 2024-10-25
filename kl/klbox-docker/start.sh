#!/usr/bin/env bash
# shellcheck source=/dev/null
set -o errexit
set -o pipefail

trap "echo kloudlite-entrypoint:CRASHED >&2" EXIT SIGINT SIGTERM

#/nix-installer/install

export IN_DEV_BOX="true"
export KL_WORKSPACE="$KL_WORKSPACE"
export KL_TMP_PATH="/kl-tmp"

cat <<EOL >/kl-tmp/global-profile
export SSH_PORT=$SSH_PORT
export IN_DEV_BOX="true"
export KL_WORKSPACE="$KL_WORKSPACE"
export KL_BASE_URL="$KL_BASE_URL"
export MAIN_PATH=$PATH
export KL_TMP_PATH="/kl-tmp"
export KLCONFIG_PATH="$KLCONFIG_PATH"
export PLATFORM_ARCH=$(uname -m)
export KL_HOST_USER="$KL_HOST_USER"
EOL

chown -R kl /kl-tmp/global-profile

mkdir -p /etc/wireguard
echo $KL_TEAM_NAME
set -x
CLUSTER_IP_RANGE=$(echo $CLUSTER_IP_RANGE | sed 's/\//###/g')
cat /.cache/kl/kl-workspace-wg.conf | sed "s/#CLUSTER_GATEWAY_IP/${CLUSTER_GATEWAY_IP:-null}/" | sed "s/#CLUSTER_IP_RANGE/${CLUSTER_IP_RANGE:-null}/" >/tmp/wg-cong
sed -i "s/###/\//" /tmp/wg-cong
set +x
sudo cp /tmp/wg-cong /etc/wireguard/kl-workspace-wg.conf
rm /tmp/wg-cong
cat /.cache/kl/vpn/${KL_TEAM_NAME}.json | jq -r .wg | base64 -d >/tmp/kl-vpn.conf
sudo cp /tmp/kl-vpn.conf /etc/wireguard/kl-vpn.conf
rm /tmp/kl-vpn.conf
sudo wg-quick up kl-vpn
sudo wg-quick up kl-workspace-wg

entrypoint_executed="/home/kl/.kloudlite_entrypoint_executed"
if [ ! -f "$entrypoint_executed" ]; then
  mkdir -p /home/kl/.config
  cp /tmp/.zshrc /home/kl/.zshrc
  cp /tmp/.bashrc /home/kl/.bashrc
  cp /tmp/.profile /home/kl/.profile
  cp /tmp/.check-online /home/kl/.check-online
  ln -sf /home/kl/.profile /home/kl/.zprofile
  cp /tmp/aliasrc /home/kl/.config/aliasrc
  echo "successfully initialized .profile and .bashrc" >>$entrypoint_executed
  # ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa -N "" <<<y >/dev/null 2>&1
fi

chown -R kl /home/kl/
sudo -u kl mkdir -p /home/kl/.config/nix
sudo -u kl echo "experimental-features = nix-command flakes" >/home/kl/.config/nix/nix.conf

# shift

export PATH=$PATH:$HOME/.local/state/nix/profiles/profile/bin

echo "kloudlite-entrypoint:INSTALLING_PACKAGES"
cat $KL_HASH_FILE | jq '.hash' -r >/tmp/hash
cat $KL_HASH_FILE | jq '.config.env | to_entries | map_values(. = "export \(.key)=\"\(.value)\"")|.[]' -r >>/tmp/env
cat >/tmp/mount.sh <<EOF
set -o errexit
set -o pipefail
vmounts=$(cat $KL_HASH_FILE | jq '.config.mounts | length')
if [ \$vmounts -gt 0 ]; then
  eval $(cat $KL_HASH_FILE | jq '.config.mounts | to_entries | map_values(. = "mkdir -p $(dirname \(.key))") | .[]' -r)
  eval $(cat $KL_HASH_FILE | jq '.config.mounts | to_entries | map_values(. = "echo \"\(.value)\" | base64 -d > \(.key)") | .[]' -r)
fi
EOF
sudo bash /tmp/mount.sh

cat >/tmp/pkg-install.sh <<EOF
set -o errexit
set -o pipefail
npkgs=$(cat $KL_HASH_FILE | jq '.config.packageHashes | length')
if [ \$npkgs -gt 0 ]; then
  export PATH=$PATH:/home/kl/.local/state/nix/profiles/profile/bin
  nix shell --log-format bar-with-logs $(cat $KL_HASH_FILE | jq '.config.packageHashes | to_entries | map_values(. = .value) | .[]' -r | xargs -I{} printf "%s " {}) --command echo "successfully installed packages"
  npath=$(nix shell $(cat $KL_HASH_FILE | jq '.config.packageHashes | to_entries | map_values(. = .value) | .[]' -r | xargs -I{} printf "%s " {}) --command printenv PATH)
  echo export PATH=$PATH:\$npath >> /kl-tmp/env
fi
EOF

sudo -u kl bash /tmp/pkg-install.sh

#nix shell $(cat $KL_HASH_FILE | jq '.config.packageHashes | to_entries | map_values(. = .value) | .[]' -r | xargs -I{} printf "%s " {}) --command echo "successfully installed packages"
#echo export PATH=$PATH:$(eval nix shell $(cat $KL_HASH_FILE | jq '.config.packageHashes | to_entries | map_values(. = .value) | .[]' -r | xargs -I{} printf "%s " {}) --command printenv PATH) >> /tmp/env
echo "export KL_HASH_FILE=$KL_HASH_FILE" >>/kl-tmp/env
echo "kloudlite-entrypoint:INSTALLING_PACKAGES_DONE"

source /kl-tmp/env

RESOLV_FILE="/etc/resolv.conf"
# add search domain to resolv.conf
# echo "search $KL_SEARCH_DOMAIN" > $RESOLV_FILE
# echo "options ndots:0" >> $RESOLV_FILE
sudo sh -c "echo \"search $KL_SEARCH_DOMAIN\" >> $RESOLV_FILE"

# sudo cp /tmp/resolv.conf /etc/resolv.conf

#if [ -d "/tmp/ssh2" ]; then
#  mkdir -p /home/kl/.ssh
#  cp /tmp/ssh2/authorized_keys /home/kl/.ssh/authorized_keys
#  cp /tmp/ssh2/id_rsa /home/kl/.ssh/id_rsa
#  cp /tmp/ssh2/id_rsa.pub /home/kl/.ssh/id_rsa.pub
#  chmod 600 /home/kl/.ssh/authorized_keys
#  echo "successfully copied ssh credentials"
#fi

bash ~/.check-online >/dev/null 2>&1 &

trap - EXIT SIGTERM SIGINT
echo "kloudlite-entrypoint:SETUP_COMPLETE"
