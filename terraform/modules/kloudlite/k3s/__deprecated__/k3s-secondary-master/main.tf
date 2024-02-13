locals {
  node_taints = flatten([
    for taint in var.node_taints : [
      "${taint.key}=${taint.value}:${taint.effect}",
    ]
  ])
}

resource "ssh_resource" "setup_k3s_on_secondary_masters" {
  host        = var.ssh_params.public_ip
  user        = var.ssh_params.user
  private_key = var.ssh_params.private_key

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
  publicIP: ${var.ssh_params.public_ip}
  serverIP: ${var.primary_master_public_ip}
  token: ${var.k3s_token}
  nodeName: ${var.node_name}
  labels: ${jsonencode(var.node_labels)}
  taints: ${jsonencode(local.node_taints)}
  SANs: ${jsonencode([var.public_dns_hostname, var.ssh_params.public_ip])}
  extraServerArgs: ${jsonencode(concat([
    "--disable-helm-controller",
    "--disable", "traefik",
    "--disable", "servicelb",
    "--node-external-ip", var.ssh_params.public_ip,
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
      "--etcd-snapshot-schedule-cron",  var.backup_to_s3.crontab_schedule,
  ] : []
))}
EOF2

sudo ln -sf $PWD/runner-config.yml /runner-config.yml
EOT
  ]
}