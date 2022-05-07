#!/bin/bash

cat > /root/.ssh/authorized_keys << EOF
${pubkey}
EOF

cat > /root/wg-ip << EOF
${wg_ip}
EOF
