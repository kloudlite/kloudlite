#! /usr/bin/env bash

## terraform params
K3S_SERVER_HOST='${k3s_server_host}'
K3S_AGENT_TOKEN='${k3s_agent_token}'
K3S_VERSION="${k3s_version}"
NODE_NAME="${node_name}"
## --tf params:END

count=60
while [ "$count" -gt 0 ]; do
  DEVICE_NAME=$(lsblk -OdJ | jq '.blockdevices[] | select(.ptuuid == null) | .name' -r)
  if [ -n "$DEVICE_NAME" ]; then
    if ! blkid /dev/"$DEVICE_NAME"; then
      echo "No filesystem found. Formatting with XFS..."
      mkfs.xfs -f "/dev/$DEVICE_NAME"
    fi

    # Mount the device if not already mounted
    if ! mount | grep -q "/dev/$DEVICE_NAME"; then
      mountpoint="/external-volume"
      mkdir -p "$mountpoint"
      sudo chown 1000:1000 -R $mountpoint
      mount -t xfs "/dev/$DEVICE_NAME" "$mountpoint"

      # Check if user-home folder exists, if not create it
      user_home="$mountpoint/user-home"
      if [ ! -d "$user_home" ]; then
        mkdir -p "$user_home"
        chown 1000:1000 -R "$user_home"
        echo "Created user-home directory at $user_home"
      fi

      # Check if nix directory exists, if not create it
      nixdir="$mountpoint/nix"
      if [ ! -d "$nixdir" ]; then
        mkdir -p "$nixdir"
        chown 1000:1000 -R "$nixdir"
        echo "Created nix directory at $nixdir"
      fi

      # sleep 3
      # DEVICE_UUID=$(lsblk -OdJ | jq '.blockdevices[] | select(.name == "$DEVICE_NAME") | .uuid' -r)
      # echo "UUID=$DEVICE_UUID $mountpoint xfs defaults,nofail 0 2" | sudo tee -a /etc/fstab
      echo "/dev/$DEVICE_NAME $mountpoint xfs defaults,nofail 0 2" | sudo tee -a /etc/fstab
    else
      echo "/dev/$DEVICE_NAME is already mounted."
    fi
    break
  fi
  count=$((count - 1))
  sleep 1
done

KLOUDLITE_CONFIG_DIRECTORY=/etc/kloudlite

debug() {
  echo "[#] $*" | tee -a "$KLOUDLITE_CONFIG_DIRECTORY/execution.log"
}

debug "ensure kloudlite config directory ($KLOUDLITE_CONFIG_DIRECTORY) exists"
mkdir -p "$KLOUDLITE_CONFIG_DIRECTORY"

debug "################# execution started at $(date) ######################"
[ $EUID -ne 0 ] && debug "this script must be run as root. current EUID is $EUID" && exit 1

create_k3s_config_file() {
  cat >"$KLOUDLITE_CONFIG_DIRECTORY/k3s.yaml" <<EOF
server: "https://$K3S_SERVER_HOST:6443"
token: "$K3S_AGENT_TOKEN"
node-name: "$NODE_NAME"
node-label:
  - "kloudlite.io/node.role=worker"
node-taint:
  - kloudlite.io/workmachine.name=$NODE_NAME:NoExecute
kubelet-arg:
  - "system-reserved=cpu=50m,memory=50Mi,ephemeral-storage=2Gi"
  - "kube-reserved=cpu=100m,memory=256Mi"
  - "eviction-hard=nodefs.available<5%,nodefs.inodesFree<5%,imagefs.available<5%"
EOF

  mkdir -p /etc/rancher/k3s
  ln -sf $KLOUDLITE_CONFIG_DIRECTORY/k3s.yaml /etc/rancher/k3s/config.yaml
}

install_k3s() {
  debug "installing k3s"
  export INSTALL_K3S_CHANNEL="stable"
  export INSTALL_K3S_SKIP_SELINUX_RPM="true"

  if [ -n "$K3S_VERSION" ]; then
    export INSTALL_K3S_VERSION="$K3S_VERSION"
  fi
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="agent" sh -
}

create_k3s_config_file
install_k3s

debug "################# execution finished at $(date) ######################"
