#!/usr/bin/env bash
set -euo pipefail

# AWS CLI Script: Create EC2 Instance for WorkMachine
# Outputs JSON to ConfigMap for controller to parse
# Can be tested locally: ./create-instance.sh

# Required environment variables:
# - MACHINE_NAME: Name of the WorkMachine
# - OWNER: Owner of the machine
# - REGION: AWS region
# - AMI: AMI ID to use
# - INSTANCE_TYPE: EC2 instance type
# - SUBNET_ID: Subnet ID
# - VPC_ID: VPC ID
# - ALLOWED_CIDR: CIDR to allow to workmachine node (default: 0.0.0.0/0)
# - ROOT_VOLUME_SIZE: Root volume size in GB
# - ROOT_VOLUME_TYPE: Root volume type (gp3, gp2, etc)
# - K3S_TOKEN: K3s agent token for joining cluster
# - K3S_SERVER_URL: K3s server URL
# - AVAILABILITY_ZONE: (optional) Availability zone
# - IAM_INSTANCE_PROFILE: (optional) IAM instance profile name

function log() {
  echo "[$(date -Iseconds)] $*"
}

function valid() {
  [ -n "$1" ] && [ "$1" != "None" ]
}

# Required variables
required_vars=(
  "MACHINE_NAME"
  "OWNER"
  # "SUBNET_ID"               # (optional)
  # "VPC_ID"                  # (default: via IMDSv2 API)
  # "REGION"                  # (default: via IMDSv2 API)
  "AMI"
  "INSTANCE_TYPE"
  # "ALLOWED_CIDR"            # (default: 0.0.0.0/0)
  "ROOT_VOLUME_SIZE"
  # "ROOT_VOLUME_TYPE"        # (default: gp3)
  "K3S_TOKEN"
  "K3S_SERVER_URL"

  # "AVAILABILITY_ZONE"       # (optional)
  # "IAM_INSTANCE_PROFILE"    # (optional)
)

for var in "${required_vars[@]}"; do
  if [ -z "${!var:-}" ]; then
    log "ERROR: Required environment variable not set ($var)"
    exit 1
  fi
done

ALLOWED_CIDR=${ALLOWED_CIDR:-0.0.0.0/0}
ROOT_VOLUME_TYPE=${ROOT_VOLUME_TYPE:-gp3}

log "Starting instance creation for ${MACHINE_NAME}"

aws_metadata_token=$(curl -sX PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")

if ! valid "${REGION:-}"; then
  log "missing REGION, fetching from instance metadata v2 API"
  REGION=$(curl -sH "X-aws-ec2-metadata-token: $aws_metadata_token" http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r '.region')
fi

if ! valid "${VPC_ID:-}"; then
  log "missing VPC_ID, fetching from instance metadata v2 API"
  mac_addr=$(curl -sH "X-aws-ec2-metadata-token: $aws_metadata_token" http://169.254.169.254/latest/meta-data/mac)
  VPC_ID=$(curl -sH "X-aws-ec2-metadata-token: $aws_metadata_token" http://169.254.169.254/latest/meta-data/network/interfaces/macs/"${mac_addr}"/vpc-id)
fi

# Step 1: Ensure SSH key pair exists for debugging
ssh_key_name="kloudlite-workmachine-ssh"

log "Checking for SSH key pair ${ssh_key_name}..."

ssh_key_exists=$(aws ec2 describe-key-pairs \
  --region "${REGION}" \
  --filters "Name=key-name,Values=${ssh_key_name}" \
  --query 'KeyPairs[0].KeyName' \
  --output text 2>/dev/null || echo "None")

if ! valid "$ssh_key_exists"; then
  log "Creating SSH key pair ${ssh_key_name}..."
  aws ec2 create-key-pair \
    --region "${REGION}" \
    --key-name "${ssh_key_name}" \
    --query 'KeyMaterial' \
    --output text

  log "Created SSH key pair: ${ssh_key_name}"
fi

log "Using SSH key pair: ${ssh_key_name}"

# Step 2: Get or create security group
log "Checking for existing security group..."

SG_NAME="kloudlite-workmachine-${MACHINE_NAME}"

SG_ID=$(aws ec2 describe-security-groups \
  --region "${REGION}" \
  --filters "Name=group-name,Values=${SG_NAME}" "Name=vpc-id,Values=${VPC_ID}" \
  --query 'SecurityGroups[0].GroupId' \
  --output text 2>/dev/null)

if ! valid "$SG_ID"; then
  log "Creating Security Group ${SG_NAME}..."
  SG_ID=$(aws ec2 create-security-group \
    --region "${REGION}" \
    --group-name "${SG_NAME}" \
    --description "Security group for WorkMachine ${MACHINE_NAME} (owner: ${OWNER})" \
    --vpc-id "${VPC_ID}" \
    --tag-specifications "ResourceType=security-group,Tags=[{Key=Name,Value=${SG_NAME}},{Key=kloudlite.io/workmachine,Value=${MACHINE_NAME}},{Key=kloudlite.io/owner,Value=${OWNER}}]" \
    --query 'GroupId' \
    --output text)

  log "Created Security Group: ${SG_ID}"

  # Add ingress rules for HTTPS (443) and SSH (22)
  log "Adding ingress rules..."
  aws ec2 authorize-security-group-ingress \
    --region "${REGION}" \
    --group-id "${SG_ID}" \
    --ip-permissions \
    "IpProtocol=tcp,FromPort=443,ToPort=443,IpRanges=[{CidrIp=${ALLOWED_CIDR},Description=HTTPS access}]" \
    "IpProtocol=tcp,FromPort=22,ToPort=22,IpRanges=[{CidrIp=${ALLOWED_CIDR},Description=SSH access}]" \
    2>/dev/null || log "Ingress rules may already exist"
fi

log "Using security group: ${SG_ID}"

# Step 3: Generate K3s user data
log "Generating K3s user data..."

USER_DATA=$(
  cat <<EOF | base64 -w 0
#!/usr/bin/env bash
set -x

# Log to both console and file
exec > >(tee /var/log/user-data.log)
exec 2>&1

echo "User-data script starting at \$(date)"

# Install K3s agent with node configuration
curl -sfL https://get.k3s.io | \\
  K3S_URL="${K3S_SERVER_URL}" \\
  K3S_TOKEN="${K3S_TOKEN}" \\
  INSTALL_K3S_EXEC="agent \\
    --node-name=${MACHINE_NAME} \\
    --node-label=kloudlite.io/workmachine=${MACHINE_NAME} \\
    --node-label=kloudlite.io/owner=${OWNER} \\
    --node-taint=kloudlite.io/workmachine=${MACHINE_NAME}:NoSchedule" \\
  sh -

echo "K3s agent installed, node joining cluster..."
EOF
)

# Step 4: Build run-instances command
log "Creating EC2 Instance..."

# Dynamically detect the AMI's root device name
ROOT_DEVICE_NAME=$(aws ec2 describe-images \
  --region "${REGION}" \
  --image-ids "${AMI}" \
  --query 'Images[0].RootDeviceName' \
  --output text)

RUN_INSTANCE_ARGS=(
  --user-data "${USER_DATA}"
  --region "${REGION}"
  --image-id "${AMI}"
  --instance-type "${INSTANCE_TYPE}"
  --security-group-ids "${SG_ID}"
  --key-name "${ssh_key_name}"
  --block-device-mappings "[{\"DeviceName\":\"${ROOT_DEVICE_NAME}\",\"Ebs\":{\"VolumeSize\":${ROOT_VOLUME_SIZE},\"VolumeType\":\"${ROOT_VOLUME_TYPE}\",\"DeleteOnTermination\":true}}]"
  --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=workmachine-${MACHINE_NAME}},{Key=kloudlite.io/workmachine,Value=${MACHINE_NAME}},{Key=kloudlite.io/owner,Value=${OWNER}},{Key=kloudlite.io/managed-by,Value=kloudlite-controller}]" "ResourceType=volume,Tags=[{Key=Name,Value=workmachine-${MACHINE_NAME}},{Key=kloudlite.io/workmachine,Value=${MACHINE_NAME}},{Key=kloudlite.io/owner,Value=${OWNER}}]"
)

valid "${SUBNET_ID:-}" &&
  RUN_INSTANCE_ARGS+=(--subnet-id "${SUBNET_ID}")

# Add optional parameters
valid "${AVAILABILITY_ZONE:-}" &&
  RUN_INSTANCE_ARGS+=(--placement "AvailabilityZone=${AVAILABILITY_ZONE}")

valid "${IAM_INSTANCE_PROFILE:-}" &&
  RUN_INSTANCE_ARGS+=(--iam-instance-profile "Name=${IAM_INSTANCE_PROFILE}")

# Create instance
INSTANCE_OUTPUT=$(aws ec2 run-instances "${RUN_INSTANCE_ARGS[@]}")

INSTANCE_ID=$(echo "$INSTANCE_OUTPUT" | jq -r '.Instances[0].InstanceId')
PRIVATE_IP=$(echo "$INSTANCE_OUTPUT" | jq -r '.Instances[0].PrivateIpAddress // ""')
AZ=$(echo "$INSTANCE_OUTPUT" | jq -r '.Instances[0].Placement.AvailabilityZone')
STATE=$(echo "$INSTANCE_OUTPUT" | jq -r '.Instances[0].State.Name')

log "Instance created: ${INSTANCE_ID} (state: ${STATE})"

# Step 5: Wait for public IP (instances take a moment to get public IP)
log "Waiting for public IP assignment..."
PUBLIC_IP=""
for i in {1..30}; do
  PUBLIC_IP=$(aws ec2 describe-instances \
    --region "${REGION}" \
    --instance-ids "${INSTANCE_ID}" \
    --query 'Reservations[0].Instances[0].PublicIpAddress' \
    --output text 2>/dev/null || echo "")

  if [ -n "$PUBLIC_IP" ] && [ "$PUBLIC_IP" != "None" ]; then
    log "Public IP assigned: ${PUBLIC_IP}"
    break
  fi

  sleep 2
done

log "Instance creation complete!"

cat >"$MACHINE_NAME.json" <<EOF
{
  "instance_id": "${INSTANCE_ID}",
  "security_group_id": "${SG_ID}",
  "public_ip": "${PUBLIC_IP:-}",
  "private_ip": "${PRIVATE_IP}",
  "region": "${REGION}",
  "availability_zone": "${AZ}",
  "state": "${STATE}",
  "ssh_key_name": "${ssh_key_name}",
  "message": "Instance created successfully. SSH: ssh -i ~/.ssh/${ssh_key_name}.pem ec2-user@${PUBLIC_IP}"
}
EOF

if command -v kubectl &>/dev/null; then
  # Running in Kubernetes Job
  kubectl create configmap "workmachine-${MACHINE_NAME}-output" \
    --from-file="$MACHINE_NAME.json" \
    --dry-run='client' -o yaml | kubectl apply -f -
  log "ConfigMap created: workmachine-${MACHINE_NAME}-output"
fi

log "Printing output for audit logs:"
jq <"$MACHINE_NAME.json"
