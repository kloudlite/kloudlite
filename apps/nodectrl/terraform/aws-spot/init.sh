#!/bin/bash

mkdir /k3s
cat >> /k3s/data.yaml << EOF
${nodeConfigYaml}
EOF

cat >> /root/.ssh/authorized_keys << EOF
${pubkey}
EOF
