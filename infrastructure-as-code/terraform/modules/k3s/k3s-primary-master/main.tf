resource "random_password" "k3s_token" {
  length  = 64
  special = false
}

locals {
  node_taints = flatten([
    for k, taint in var.node_taints : [
      "--node-taint", "${k}=${taint.value}:${taint.effect}",
    ]
  ])
}

resource "ssh_resource" "setup_k3s_on_primary_master" {
  host        = var.public_ip
  user        = var.ssh_params.user
  private_key = var.ssh_params.private_key

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
  publicIP: ${var.public_ip}
  token: ${random_password.k3s_token.result}
  nodeName: ${var.node_name}
  labels: ${jsonencode(var.node_labels)}
  SANs: ${jsonencode(concat([var.public_dns_hostname], var.k3s_master_nodes_public_ips))}
  extraServerArgs: ${jsonencode(concat([
    "--disable-helm-controller",
    "--disable", "traefik",
    "--disable", "servicelb",
    "--node-external-ip", var.public_ip,
    "--tls-san-security",
    "--flannel-external-ip",
  ],
  length(var.node_taints) >  0 ? local.node_taints : [],

  var.backup_to_s3.enabled ? [
      "--etcd-s3",
      "--etcd-s3-endpoint", "s3.amazonaws.com",
      "--etcd-s3-access-key", var.backup_to_s3.aws_access_key,
      "--etcd-s3-secret-key", var.backup_to_s3.aws_secret_key,
      "--etcd-s3-bucket", var.backup_to_s3.bucket_name,
      "--etcd-s3-region", var.backup_to_s3.bucket_region,
      "--etcd-s3-folder", var.backup_to_s3.bucket_folder,
      "--etcd-snapshot-compress",
      "--etcd-snapshot-schedule-cron",  var.backup_to_s3.cron_schedule,
  ] : [],

  var.restore_from_latest_s3_snapshot ? [
      "--cluster-reset",
      "--cluster-reset-restore-path", "$latest_snapshot",
  ]:  []
))}

EOF2

    sudo ln -sf $PWD/runner-config.yml /runner-config.yml
EOT
  ]
}

resource "null_resource" "wait_till_k3s_server_is_ready" {
  connection {
    type        = "ssh"
    host        = var.public_ip
    user        = var.ssh_params.user
    private_key = var.ssh_params.private_key
  }

  depends_on = [ssh_resource.setup_k3s_on_primary_master]

  provisioner "remote-exec" {
    inline = [
      <<EOC
    chmod +x ./k8s-user-account.sh
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

    ./k8s-user-account.sh kubeconfig.yml
EOC
    ]
  }
}

resource "ssh_resource" "copy_kubeconfig" {
  host        = var.public_ip
  user        = var.ssh_params.user
  private_key = var.ssh_params.private_key

  depends_on = [null_resource.wait_till_k3s_server_is_ready]

  timeout     = "10s"
  retry_delay = "2s"

  when = "create"

  commands = [
    <<EOT
cat kubeconfig.yml | base64 | tr -d '\n'
EOT
  ]
}

