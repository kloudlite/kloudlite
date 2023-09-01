{
  "access_key": "${AWS_ACCESS_KEY_ID}",
  "secret_key": "${AWS_SECRET_ACCESS_KEY}",
  "region": "${REGION}",
  "node_name": "",
  "instance_type": "c6a.large",
  "public_key_path": "${PUBLIC_KEY_PATH}",
  "private_key_path": "${PRIVATE_KEY_PATH}",
  "keys_path": "./",
  "ami": "",

  "k3s_token": "${K3S_TOKEN}",

  "master_nodes_config": {
    "name": "${MASTER_NODE_PREFIX}",
    "instance_type": "${INSTANCE_TYPE}",
    "count": 3,
    "availability_zones": "${AZ}",
    "ami": "${AMI}"
  },

  "worker_nodes_config": {
    "name": "${WORKER_NODE_PREFIX}",
    "instance_type": "${INSTANCE_TYPE}",
    "count": 1,
    "availability_zones": "${AZ}",
    "ami": "${AMI}"
  },

  "cloudflare": {
    "api_token": "${CLOUDFLARE_API_TOKEN}",
    "zone_id": "${CLOUDFLARE_ZONE_ID}",
    "domain": "${CLOUDFLARE_DOMAIN}"
  }
}
