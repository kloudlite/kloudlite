#!/usr/bin/env bash

# AWS CLI Script: Resize Root Volume
# Uses hybrid approach: INSTANCE_ID env var OR query by tags

# Required environment variables:
# - INSTANCE_ID: Instance ID
# - TARGET_VOLUME_SIZE: Updated Disk Volume Size

function log() {
  echo "[$(date -Iseconds)] $*"
}

# Required variables
required_vars=(
  "INSTANCE_ID"
  "TARGET_VOLUME_SIZE"
)

for var in "${required_vars[@]}"; do
  if [ -z "${!var:-}" ]; then
    log "ERROR: Required environment variables not set: ($var)"
    exit 1
  fi
done

aws ec2 reboot-instances --instance-ids i-xxxxxxxxxxxxxxxxx
