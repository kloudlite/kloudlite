#!/bin/bash

apt install -y wireguard

cat > /root/install.sh <<EOF
#!/bin/bash
if ! command -v k3s &> /dev/null
then
  curl -sfL https://get.k3s.io |  K3S_TOKEN=$1 K3S_NODE_NAME=master INSTALL_K3S_EXEC="server --cluster-init --datastore-endpoint='$2'  --disable traefik" sh -s -
fi
EOF

cat > /root/join-agent.sh <<EOF
#!/bin/bash
if ! command -v k3s &> /dev/null
then
  curl -sfL https://get.k3s.io |  K3S_TOKEN=$1 K3S_NODE_NAME=$3  INSTALL_K3S_EXEC="agent --server https://$2:6443" sh -s -
fi
EOF

cat > /root/join-master.sh <<EOF
#!/bin/bash
if ! command -v k3s &> /dev/null
then
  curl -sfL https://get.k3s.io |  K3S_TOKEN=$1 K3S_NODE_NAME=$3  INSTALL_K3S_EXEC="server --datastore-endpoint=$4  --server https://$2:6443" sh -s -
fi
EOF



cat > /root/.ssh/authorized_keys << EOF
${pubkey}
EOF

cat > /root/wg-ip << EOF
${wg_ip}
EOF
