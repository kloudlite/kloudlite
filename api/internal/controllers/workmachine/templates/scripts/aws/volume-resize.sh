#!/bin/bash
set -euo pipefail

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

log "Starting Root Volume Resize for instance ($INSTANCE_ID)"

root_device_name=$(
  aws ec2 describe-instances \
    --instance-id "${INSTANCE_ID}" \
    --query "Reservations[0].Instances[0].RootDeviceName" \
    --output text
)

log "found, root device name: $root_device_name"

volume_id=$(
  aws ec2 describe-instances \
    --instance-id "${INSTANCE_ID}" \
    --query "Reservations[].Instances[].BlockDeviceMappings[?DeviceName=='${root_device_name}'].Ebs.VolumeId" \
    --output text
)

log "found, root volume id: $volume_id"

aws ec2 describe-volumes \
  --volume-id "${volume_id}" \
  --query "Volumes[].{Size:Size,State:State}" \
  --output table

log "Resizing Volume to ${TARGET_VOLUME_SIZE}"

aws ec2 modify-volume \
  --volume-id "${volume_id}" \
  --size "${TARGET_VOLUME_SIZE}"

log "Stopping Instance for volume resize to go through"
source stop-instance.sh

log "Starting Instance back"
source start-instance.sh
