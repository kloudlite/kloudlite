locals {
  k3s_masters_private = [
    {
      description = "k3s HA masters: etcd communication, source: https://docs.k3s.io/installation/requirements#networking"
      protocol    = "tcp"
      ports       = ["2379"]
    },

    {
      description = "k3s HA masters: etcd communication, source: https://docs.k3s.io/installation/requirements#networking"
      protocol    = "tcp"
      ports       = ["2380"]
    },
  ]

  k3s_masters_public = [
    {
      description = "k8s api server communication, source: https://docs.k3s.io/installation/requirements#networking"
      protocol    = "tcp"
      ports       = ["6443"]
    },
    {
      description = "k3s masters: flannel wireguard_native communication, source: https://docs.k3s.io/installation/requirements#networking"
      protocol    = "udp"
      ports       = ["51820"]
    },
  ]

  metrics_server = [
    {
      description = "allowing metrics server communication, source: https://docs.k3s.io/installation/requirements#networking"
      protocol    = "tcp"
      ports       = ["10250"]
    },
  ]

  incoming_http_traffic = [
    {
      description = "allows http communication over port 80"
      protocol    = "tcp"
      ports       = ["80"]
    },

    {
      description = "allows https communication over port 443"
      protocol    = "tcp"
      ports       = ["443"]
    },
  ]

  incoming_ssh_connection = [
    {
      description = "allows ssh connect"
      protocol    = "tcp"
      ports       = ["22"]
    },
  ]

  incoming_dns_traffic = [
    {
      description = "allows dns communication"
      protocol    = "udp"
      ports       = ["53"]
    },
  ]

  node_ports = [
    {
      description = "open node ports"
      protocol    = "tcp"
      ports       = ["30000-32768"]
    },
    {
      description = "open node ports"
      protocol    = "udp"
      ports       = ["30000-32768"]
    },
  ]
}

resource "google_compute_firewall" "k3s_master_nodes_private" {
  count = var.for_master_nodes ? 1 : 0

  // INFO: name must be less than 63 chars
  name    = "${var.name_prefix}-master-priv"
  network = var.network_name

  dynamic "allow" {
    for_each = { for k, v in local.k3s_masters_private : k => v }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  dynamic "allow" {
    for_each = { for k, v in local.metrics_server : k => v }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  // Target tags can be used to apply this rule to specific instances
  target_tags = var.target_tags

  source_tags = var.target_tags
}

resource "google_compute_firewall" "k3s_master_nodes_public" {
  count = var.for_master_nodes ? 1 : 0

  name    = "${var.name_prefix}-master-pub"
  network = var.network_name

  dynamic "allow" {
    for_each = { for k, v in local.k3s_masters_public : k => v }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  dynamic "allow" {
    for_each = { for k, v in local.incoming_http_traffic : k => v if var.allow_incoming_http_traffic }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  dynamic "allow" {
    for_each = { for k, v in local.node_ports : k => v if var.allow_node_ports }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  dynamic "allow" {
    for_each = { for k, v in local.incoming_ssh_connection : k => v if var.allow_ssh }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  dynamic "allow" {
    for_each = { for k, v in local.incoming_dns_traffic : k => v if var.allow_dns_traffic }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  // Target tags can be used to apply this rule to specific instances
  target_tags = var.target_tags

  // This specifies the source ranges that are allowed to access the instances
  // 0.0.0.0/0 allows access from any IP address. Adjust as necessary for your security requirements.
  source_ranges = ["0.0.0.0/0"]
}

resource "google_compute_firewall" "k3s_worker_nodes" {
  count = var.for_worker_nodes ? 1 : 0

  name    = "${var.name_prefix}-worker-pub"
  network = var.network_name

  dynamic "allow" {
    for_each = { for k, v in local.incoming_http_traffic : k => v if var.allow_incoming_http_traffic }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  dynamic "allow" {
    for_each = { for k, v in local.node_ports : k => v if var.allow_node_ports }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  dynamic "deny" {
    for_each = { for k, v in local.incoming_ssh_connection : k => v }
    content {
      protocol = deny.value.protocol
      ports    = deny.value.ports
    }
  }

  // Target tags can be used to apply this rule to specific instances
  target_tags = var.target_tags

  // This specifies the source ranges that are allowed to access the instances
  // 0.0.0.0/0 allows access from any IP address. Adjust as necessary for your security requirements.
  source_ranges = ["0.0.0.0/0"]
}

resource "google_compute_firewall" "vm_group" {
  count = var.for_vm_group ? 1 : 0

  name    = "${var.name_prefix}-vm"
  network = var.network_name

  dynamic "allow" {
    for_each = { for k, v in local.incoming_http_traffic : k => v if var.allow_incoming_http_traffic }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  // this allow/deny pair is done to ensure at least one pair is always there

  dynamic "allow" {
    for_each = { for k, v in local.incoming_ssh_connection : k => v if var.allow_ssh }
    content {
      protocol = allow.value.protocol
      ports    = allow.value.ports
    }
  }

  dynamic "deny" {
    for_each = { for k, v in local.incoming_ssh_connection : k => v if !var.allow_ssh }
    content {
      protocol = deny.value.protocol
      ports    = deny.value.ports
    }
  }

  // Target tags can be used to apply this rule to specific instances
  target_tags = var.target_tags

  // This specifies the source ranges that are allowed to access the instances
  // 0.0.0.0/0 allows access from any IP address. Adjust as necessary for your security requirements.
  source_ranges = ["0.0.0.0/0"]
}
