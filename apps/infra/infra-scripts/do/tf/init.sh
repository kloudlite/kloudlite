#!/bin/bash

cat >> /root/.ssh/authorized_keys << EOF
${pubkey}
EOF
