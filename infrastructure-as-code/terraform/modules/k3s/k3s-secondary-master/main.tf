locals {
  node_taints = {
    for k, v  in var.secondary_masters : k => flatten([
      for tk, taint in v.node_taints : [
        "--node-taint", "${tk}=${taint.value}:${taint.effect}",
      ]
    ])
  }
}

resource "null_resource" "setup_k3s_on_secondary_masters" {
  for_each = {for idx, config in var.secondary_masters : idx => config}
  connection {
    type        = "ssh"
    user        = each.value.ssh_params.user
    host        = each.value.public_ip
    private_key = each.value.ssh_params.private_key
  }

  provisioner "remote-exec" {
    inline = [
      <<-EOC
      if [ "${var.restore_from_latest_s3_snapshot}" == "true" ]; then
        systemctl stop kloudlite-k3s.service
        rm -rf /var/lib/rancher/k3s/server/db/
      fi
      cat > runner-config.yml<<EOF2
      runAs: secondaryMaster
      secondaryMaster:
        publicIP: ${each.value.public_ip}
        serverIP: ${var.primary_master_public_ip}
        token: ${var.k3s_token}
        nodeName: ${each.key}
        labels: ${jsonencode(each.value.node_labels)}
        #SANs: ${jsonencode(concat([var.public_dns_hostname, "10.43.0.1"], var.k3s_master_nodes_public_ips))}
        SANs: ${jsonencode(concat([var.public_dns_hostname], var.k3s_master_nodes_public_ips))}
        extraServerArgs: ${jsonencode(concat([
          "--disable-helm-controller",
          "--disable", "traefik",
          "--disable", "servicelb",
          "--node-external-ip", each.value.public_ip,
          "--tls-san-security",
          "--flannel-external-ip",
        ],
        length(local.node_taints[each.key]) >  0 ? local.node_taints[each.key] : [],
        var.backup_to_s3.enabled ? [
            "--etcd-s3",
            "--etcd-s3-endpoint", "s3.amazonaws.com",
            "--etcd-s3-access-key", var.backup_to_s3.aws_access_key,
            "--etcd-s3-secret-key", var.backup_to_s3.aws_secret_key,
            "--etcd-s3-bucket", var.backup_to_s3.bucket_name,
            "--etcd-s3-region", var.backup_to_s3.bucket_region,
            "--etcd-s3-folder", var.backup_to_s3.bucket_folder,
            "--etcd-snapshot-compress",
            "--etcd-snapshot-schedule-cron",  each.value.k3s_backup_cron_schedule,
        ] : []
      ))}

      EOF2

      sudo ln -sf $PWD/runner-config.yml /runner-config.yml
      EOC
    ]
  }
}
