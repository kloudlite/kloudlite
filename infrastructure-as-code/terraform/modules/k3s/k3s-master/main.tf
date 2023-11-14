locals {
  primary_master_role   = "primary-master"
  secondary_master_role = "secondary-master"

  node_taints = flatten([
    for taint in var.node_taints : [
      "${taint.key}=${taint.value}:${taint.effect}",
    ]
  ])
}


resource "random_password" "k3s_server_token" {
  length  = 64
  special = false
}

resource "random_password" "k3s_agent_token" {
  length  = 64
  special = false
}

locals {
  backup_crontab_schedule = {
    # for idx, name in local.master_names : name => "*/1 * * * *"
    for idx, node_name in [for k, _node_cfg in var.master_nodes : k] : node_name =>
    "* ${2 * (tonumber(idx) + 1)}/${2 * length(var.master_nodes)} * * *"
  }

  k3s_server_extra_args = {
    for k, v in var.master_nodes : k=> concat(
      [
        "--agent-token", random_password.k3s_agent_token.result,
        "--disable-helm-controller",
        "--disable", "traefik",
        "--disable", "servicelb",
        "--node-external-ip", v.public_ip,
        "--tls-san-security",
        "--flannel-external-ip",
        "--cluster-domain", var.cluster_internal_dns_host,
      ],
      var.backup_to_s3.enabled ? [
        "--etcd-s3",
        "--etcd-s3-endpoint", "s3.amazonaws.com",
        "--etcd-s3-access-key", var.backup_to_s3.aws_access_key,
        "--etcd-s3-secret-key", var.backup_to_s3.aws_secret_key,
        "--etcd-s3-bucket", var.backup_to_s3.bucket_name,
        "--etcd-s3-region", var.backup_to_s3.bucket_region,
        "--etcd-s3-folder", var.backup_to_s3.bucket_folder,
        "--etcd-snapshot-compress",
        "--etcd-snapshot-schedule-cron", local.backup_crontab_schedule[k],
      ] : [],
      var.extra_server_args
    )
  }

}

resource "ssh_resource" "k3s_primary_master" {
  for_each = {for k, v in var.master_nodes : k => v if v.role == local.primary_master_role}

  host        = each.value.public_ip
  user        = var.ssh_params.user
  private_key = var.ssh_params.private_key

  depends_on = [
    random_password.k3s_server_token,
    random_password.k3s_agent_token,
  ]

  timeout     = "3m"
  retry_delay = "5s"

  when = "create"

  file {
    source      = "${path.module}/scripts/k8s-user-account.sh"
    destination = "./k8s-user-account.sh"
    permissions = 0755
  }

  commands = [
    <<-EOT

    if [ "${var.restore_from_latest_s3_snapshot}" = "true" ]; then
      cat > k3s-list-snapshots.sh <<'EOF2'

sudo k3s etcd-snapshot list \
  --s3  \
  --s3-region="${var.backup_to_s3.bucket_region}" \
  --s3-folder="${var.backup_to_s3.bucket_folder}" \
   --s3-bucket="${var.backup_to_s3.bucket_name}" \
   --s3-access-key="${var.backup_to_s3.aws_access_key}" \
   --s3-secret-key="${var.backup_to_s3.aws_secret_key}"
EOF2

      latest_snapshot=$(bash k3s-list-snapshot.sh 2> /dev/null | tail -n +2 | sort -k 3 -r | head -n +1 | awk '{print $1}')
      [ -z "$latest_snapshot" ] && echo "no snapshot found, exiting ..." && exit 1
    fi

    echo "setting up k3s on primary master"
    cat > runner-config.yml <<EOF2
runAs: primaryMaster
primaryMaster:
  publicIP: ${each.value.public_ip}
  token: ${random_password.k3s_server_token.result}
  nodeName: ${each.key}
  labels: ${jsonencode(each.value.node_labels)}
  SANs: ${jsonencode([var.public_dns_host, each.value.public_ip])}
  taints: ${jsonencode(local.node_taints)}
  extraServerArgs: ${jsonencode(concat(
    local.k3s_server_extra_args[each.key],
    var.restore_from_latest_s3_snapshot ? [
        "--cluster-reset",
        "--cluster-reset-restore-path", "$latest_snapshot",
    ]:  []
  )
)}

EOF2

    sudo ln -sf $PWD/runner-config.yml /runner-config.yml
EOT
  ]
}

locals {
  primary_master_node_name = one([
    for node_name, node_cfg in var.master_nodes : node_name
    if node_cfg.role == local.primary_master_role
  ])
}

resource "null_resource" "wait_till_k3s_primary_server_is_ready" {
  for_each = {for k, v in var.master_nodes : k => v if v.role == local.primary_master_role}

  connection {
    type        = "ssh"
    host        = each.value.public_ip
    user        = var.ssh_params.user
    private_key = var.ssh_params.private_key
  }

  depends_on = [ssh_resource.k3s_primary_master]

  provisioner "remote-exec" {
    inline = [
      <<EOC
    export KUBECTL='sudo k3s kubectl'

    echo "checking whether /etc/rancher/k3s/k3s.yaml file exists"
    while true; do
      if [ ! -f /etc/rancher/k3s/k3s.yaml ]; then
        echo 'k3s yaml not found, re-checking in 1s'
        sleep 1
        continue
      fi

      echo "/etc/rancher/k3s/k3s.yaml file found"
      break
    done

    echo "checking whether k3s server is accepting connections"
    while true; do
      lines=$($KUBECTL get nodes | wc -l)

      if [ "$lines" -lt 2 ]; then
        echo "k3s server is not accepting connections yet, retrying in 1s ..."
        sleep 1
        continue
      fi
      echo "successful, k3s server is now accepting connections"
      break
    done

    echo "[#] k3s server is now fully ready, provisioning a new revocable kubeconfig"
    chmod +x ./k8s-user-account.sh
    ./k8s-user-account.sh kubeconfig.yml
EOC
    ]
  }
}

resource "ssh_resource" "create_revocable_kubeconfig" {
  for_each = {for k, v in var.master_nodes : k => v if v.role == local.primary_master_role}

  host        = each.value.public_ip
  user        = var.ssh_params.user
  private_key = var.ssh_params.private_key

  depends_on = [null_resource.wait_till_k3s_primary_server_is_ready]

  timeout     = "20s"
  retry_delay = "2s"

  when = "create"

  commands = [
    <<EOT
cat kubeconfig.yml | base64 | tr -d '\n'
EOT
  ]
}

resource "ssh_resource" "k3s_secondary_masters" {
  for_each = {for k, v in var.master_nodes : k => v if v.role == local.secondary_master_role}

  host        = each.value.public_ip
  user        = var.ssh_params.user
  private_key = var.ssh_params.private_key

  depends_on = [ssh_resource.k3s_primary_master]

  when = "create"

  timeout     = "2m"
  retry_delay = "2s"

  commands = [
    <<EOT
if [ "${var.restore_from_latest_s3_snapshot}" == "true" ]; then
  systemctl stop kloudlite-k3s.service
  rm -rf /var/lib/rancher/k3s/server/db/
fi

cat > runner-config.yml<<EOF2
runAs: secondaryMaster
secondaryMaster:
  publicIP: ${each.value.public_ip}
  serverIP: ${var.master_nodes[local.primary_master_node_name].public_ip}
  token: ${random_password.k3s_server_token.result}
  nodeName: ${each.key}
  labels: ${jsonencode(each.value.node_labels)}
  taints: ${jsonencode(local.node_taints)}
  SANs: ${jsonencode([var.public_dns_host, each.value.public_ip])}
  extraServerArgs: ${jsonencode(local.k3s_server_extra_args[each.key])}
EOF2

sudo ln -sf $PWD/runner-config.yml /runner-config.yml
EOT
  ]
}
