#!/bin/bash
set -euo pipefail

# AWS CLI Script: Stop EC2 Instance
# Uses hybrid approach: INSTANCE_ID env var OR query by tags

# Required environment variables:
# - REGION: AWS region
# - INSTANCE_ID: Instance ID

function log() {
  echo "[$(date -Iseconds)] $*"
}

# Required variables
required_vars=(
  "INSTANCE_ID"
)

for var in "${required_vars[@]}"; do
  if [ -z "${!var:-}" ]; then
    log "ERROR: Required environment variables not set: ($var)"
    exit 1
  fi
done

log "Stopping Instance: ${INSTANCE_ID}"

aws ec2 stop-instances \
  --region "${REGION}" \
  --instance-ids "${INSTANCE_ID}"

log "Waiting till instance is fully stopped ..."

aws ec2 wait instance-stopped \
  --region "${REGION}" \
  --instance-ids "${INSTANCE_ID}"

log "Instance ($INSTANCE_ID) stopped"
