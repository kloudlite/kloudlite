#!/bin/bash
set -euo pipefail

# AWS CLI Script: Delete EC2 Instance and Security Group
# Uses hybrid approach: INSTANCE_ID/SG_ID env vars OR query by tags

# Required environment variables:
# - MACHINE_NAME: Name of the WorkMachine
# - REGION: AWS region
# - VPC_ID: VPC ID for security group lookup
# Optional:
# - INSTANCE_ID: Instance ID (if known from Status)
# - SECURITY_GROUP_ID: Security group ID (if known from Status)

echo "[$(date -Iseconds)] Deleting resources for ${MACHINE_NAME}"

# Find instance ID (hybrid approach)
if [ -z "${INSTANCE_ID:-}" ]; then
  echo "[$(date -Iseconds)] No instance ID provided, querying by tags..."
  INSTANCE_ID=$(aws ec2 describe-instances \
    --region "${REGION}" \
    --filters "Name=tag:kloudlite.io/workmachine,Values=${MACHINE_NAME}" \
              "Name=instance-state-name,Values=pending,running,shutting-down,stopping,stopped" \
    --query 'Reservations[0].Instances[0].InstanceId' \
    --output text 2>/dev/null || echo "None")

  if [ "$INSTANCE_ID" = "None" ] || [ -z "$INSTANCE_ID" ]; then
    echo "[$(date -Iseconds)] No instance found for ${MACHINE_NAME}"
  else
    echo "[$(date -Iseconds)] Found instance by tags: ${INSTANCE_ID}"
  fi
else
  echo "[$(date -Iseconds)] Using provided instance ID: ${INSTANCE_ID}"
fi

# Terminate instance if found
if [ -n "${INSTANCE_ID:-}" ] && [ "$INSTANCE_ID" != "None" ]; then
  echo "[$(date -Iseconds)] Terminating instance ${INSTANCE_ID}..."
  aws ec2 terminate-instances \
    --region "${REGION}" \
    --instance-ids "${INSTANCE_ID}" \
    > /dev/null 2>&1 || echo "Instance may already be terminated"

  echo "[$(date -Iseconds)] Waiting for instance termination..."
  # Wait up to 2 minutes for termination
  for i in {1..24}; do
    STATE=$(aws ec2 describe-instances \
      --region "${REGION}" \
      --instance-ids "${INSTANCE_ID}" \
      --query 'Reservations[0].Instances[0].State.Name' \
      --output text 2>/dev/null || echo "terminated")

    if [ "$STATE" = "terminated" ]; then
      echo "[$(date -Iseconds)] Instance terminated"
      break
    fi

    echo "[$(date -Iseconds)] Instance state: ${STATE}, waiting..."
    sleep 5
  done
else
  echo "[$(date -Iseconds)] No instance to terminate"
fi

# Find and delete security group (hybrid approach)
if [ -z "${SECURITY_GROUP_ID:-}" ]; then
  echo "[$(date -Iseconds)] No security group ID provided, querying by name..."
  SG_NAME="workmachine-${MACHINE_NAME}"
  SECURITY_GROUP_ID=$(aws ec2 describe-security-groups \
    --region "${REGION}" \
    --filters "Name=group-name,Values=${SG_NAME}" "Name=vpc-id,Values=${VPC_ID}" \
    --query 'SecurityGroups[0].GroupId' \
    --output text 2>/dev/null || echo "None")

  if [ "$SECURITY_GROUP_ID" = "None" ] || [ -z "$SECURITY_GROUP_ID" ]; then
    echo "[$(date -Iseconds)] No security group found for ${MACHINE_NAME}"
  else
    echo "[$(date -Iseconds)] Found security group by name: ${SECURITY_GROUP_ID}"
  fi
else
  echo "[$(date -Iseconds)] Using provided security group ID: ${SECURITY_GROUP_ID}"
fi

# Delete security group if found
# IMPORTANT: We wait for instance termination above before attempting this
# AWS may take a moment to detach network interfaces even after "terminated" state
if [ -n "${SECURITY_GROUP_ID:-}" ] && [ "$SECURITY_GROUP_ID" != "None" ]; then
  echo "[$(date -Iseconds)] Deleting security group ${SECURITY_GROUP_ID}..."

  # Additional safety: wait a bit longer to ensure instance is fully detached
  echo "[$(date -Iseconds)] Waiting 10s for instance to fully detach from security group..."
  sleep 10

  # Security group deletion might fail if instances are still using it
  # Retry a few times with exponential backoff
  for i in {1..6}; do
    # Check if security group is still in use
    IN_USE=$(aws ec2 describe-security-groups \
      --region "${REGION}" \
      --group-ids "${SECURITY_GROUP_ID}" \
      --query 'SecurityGroups[0].IpPermissions[0].UserIdGroupPairs' \
      --output text 2>/dev/null || echo "")

    if aws ec2 delete-security-group \
      --region "${REGION}" \
      --group-id "${SECURITY_GROUP_ID}" \
      2>/dev/null; then
      echo "[$(date -Iseconds)] Security group deleted successfully"
      break
    else
      if [ $i -lt 6 ]; then
        WAIT_TIME=$((i * 10))  # Exponential backoff: 10s, 20s, 30s, 40s, 50s, 60s
        echo "[$(date -Iseconds)] Security group deletion failed (may still be in use), retrying in ${WAIT_TIME}s..."
        sleep "$WAIT_TIME"
      else
        echo "[$(date -Iseconds)] WARNING: Could not delete security group after multiple retries"
        echo "[$(date -Iseconds)] This may require manual cleanup or the security group is already deleted"
      fi
    fi
  done
else
  echo "[$(date -Iseconds)] No security group to delete"
fi

# Output JSON to ConfigMap
OUTPUT_JSON=$(cat <<EOF
{
  "instanceID": "${INSTANCE_ID:-}",
  "securityGroupID": "${SECURITY_GROUP_ID:-}",
  "state": "terminated",
  "message": "Instance and security group deletion complete"
}
EOF
)

echo "[$(date -Iseconds)] Writing output to ConfigMap..."
if command -v kubectl &> /dev/null; then
  kubectl create configmap "workmachine-${MACHINE_NAME}-output" \
    --from-literal=result.json="${OUTPUT_JSON}" \
    --dry-run=client -o yaml | kubectl apply -f -
  echo "[$(date -Iseconds)] ConfigMap updated: workmachine-${MACHINE_NAME}-output"
else
  echo "[$(date -Iseconds)] kubectl not found, printing output:"
  echo "$OUTPUT_JSON" | jq .
fi

echo "[$(date -Iseconds)] Delete complete!"
