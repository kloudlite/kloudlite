#!/bin/bash

set -o errexit
set -o pipefail

trap "echo kloudlite-entrypoint:CRASHED >&2" EXIT SIGINT SIGTERM

entrypoint_executed="/home/kl/.kloudlite_entrypoint_executed"
if [ ! -f "$entrypoint_executed" ]; then
    cp /tmp/.bashrc /home/kl/.bashrc
    cp /tmp/.profile /home/kl/.profile
    echo "successfully initialized .profile and .bashrc" >> $entrypoint_executed
fi

shift
# Read the JSON file into a variable
echo $@ | jq -r | jq -r '.packages | join(" ")'> /home/kl/.nix-shell-args
echo "$@" | jq -r > /tmp/sample.json
chown -R kl /tmp/sample.json

echo $@ | jq -r | jq -r '.envVars | map_values(. = "export \(.key)=\(.value)") | .[]' > /home/kl/.env-vars
chown -R kl /home/kl/.env-vars

echo "kloudlite-entrypoint:INSTALLING_PACKAGES"
sudo -u kl bash -c "/home/kl/.nix-profile/bin/nix-shell -p $(cat /home/kl/.nix-shell-args) --run 'echo kloudlite-entrypoint:INSTALLING_PACKAGES_DONE'"

chown -R kl /home/kl/.nix-shell-args

if [ -d "/tmp/ssh2" ]; then
    mkdir -p /home/kl/.ssh
    chown -R kl /home/kl/.ssh
    cp /tmp/ssh2/authorized_keys /home/kl/.ssh/authorized_keys
    chmod 600 /home/kl/.ssh/authorized_keys
    chown kl /home/kl/.ssh/authorized_keys
    echo "successfully copied ssh credentials"
fi 

sudo chown -R kl /home/kl/workspace
sudo /mounter --conf /tmp/sample.json

trap - EXIT SIGTERM SIGINT
echo "kloudlite-entrypoint: SETUP_COMPLETE"
/usr/sbin/sshd -D
