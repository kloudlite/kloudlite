#!/bin/bash

cat > /root/.ssh/authorized_keys << EOF
${pubkey}
EOF

echo "curl -sfL https://get.k3s.io | sh -s - \$@" > /tmp/k3s-install.sh && chomod +x /tmp/k3s-install.sh
