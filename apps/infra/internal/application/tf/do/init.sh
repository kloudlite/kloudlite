#!/bin/bash
/root/scripts/wait-for-on.sh

cat >> /root/.ssh/authorized_keys << EOF
${pubkey}
EOF

