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

function valid() {
  [ -n "$1" ] && [ "$1" != "None" ]
}

# Validate required environment variables before starting
MISSING_VARS=()

# Required variables
required_vars=(
  "REGION"
  "INSTANCE_ID"
)

for var in "${required_vars[@]}"; do
  if [ -z "${!var:-}" ]; then
    MISSING_VARS+=("$var")
  fi
done

if [ ${#MISSING_VARS[@]} -gt 0 ]; then
  log "ERROR: Required environment variables are not set:"
  for var in "${MISSING_VARS[@]}"; do
    log "  - $var"
  done
  log ""
  log "Please set all required variables before running this script. Exiting ..."
  exit 1
fi

log "Stopping Instance: ${INSTANCE_ID}"

aws ec2 stop-instances \
  --region "${REGION}" \
  --instance-ids "${INSTANCE_ID}"

log "Waiting till instance is fully stopped ..."

aws ec2 wait instance-stopped \
  --region "${REGION}" \
  --instance-ids "${INSTANCE_ID}"
