#!/bin/bash

cat >> /root/.ssh/authorized_keys << EOF
${pubkey}
EOF

HOSTNAME=${hostname}
OLDHOSTNAME=$(cat /etc/hostname)
hostnamectl set-hostname $HOSTAME
sed -i "s/$OLDHOSTNAME/$HOSTNAME/g" /etc/hostname
sed -i "s/$OLDHOSTNAME/$HOSTNAME/g" /etc/hosts
