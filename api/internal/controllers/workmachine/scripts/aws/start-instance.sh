#!/bin/bash
set -euo pipefail

# AWS CLI Script: Start EC2 Instance
# Uses hybrid approach: INSTANCE_ID env var OR query by tags

# Required environment variables:
# - REGION: AWS region
# - INSTANCE_ID: Instance ID

function log() {
  echo "[$(date -Iseconds)] $*"
}

function valid() {
  [ -n "$1" ] && [ "$1" != "None" ]
}

# Required variables
required_vars=(
  "REGION"
  "INSTANCE_ID"
)

for var in "${required_vars[@]}"; do
  if [ -z "${!var:-}" ]; then
    log "ERROR: Required environment variables not set: ($var)"
    exit 1
  fi
done

log "Starting Instance: ${INSTANCE_ID}"
aws ec2 start-instances \
  --region "${REGION}" \
  --instance-ids "${INSTANCE_ID}"

log "Waiting till instance is in running state ..."
aws ec2 wait instance-running \
  --region "${REGION}" \
  --instance-ids "${INSTANCE_ID}"
