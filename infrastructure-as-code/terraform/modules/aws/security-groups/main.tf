locals {
  k8s_node_ports = [
    {
      description = "allowing node-port range for kubernetes services (tcp), exposed with node-port"
      from_port   = 30000
      protocol    = "tcp"
      to_port     = 32768
      cidr_blocks = ["0.0.0.0/0"]
    },

    {
      description = "allowing node-port range for kubernetes services (tcp), exposed with node-port"
      from_port   = 30000
      protocol    = "udp"
      to_port     = 32768
      cidr_blocks = ["0.0.0.0/0"]
    }
  ]

  incoming_http_traffic = [
    {
      description = "allows http communication from outside to the cluster"
      from_port   = 80
      protocol    = "tcp"
      to_port     = 80
      cidr_blocks = ["0.0.0.0/0"]
    },
    {
      description = "allows https communication from outside to the cluster"
      from_port   = 443
      protocol    = "tcp"
      to_port     = 443
      cidr_blocks = ["0.0.0.0/0"]
    },
  ]

  incoming_ssh = [
    {
      description = "allows ssh communication from outside to the cluster"
      from_port   = 22
      protocol    = "tcp"
      to_port     = 22
      cidr_blocks = ["0.0.0.0/0"]
    }
  ]

  incoming_metrics_server = [
    {
      description = "allowing metrics server communication, source: https://docs.k3s.io/installation/requirements#networking"
      from_port   = 10250
      protocol    = "tcp"
      to_port     = 10250
      cidr_blocks = ["0.0.0.0/0"]
    }
  ]

  outgoing_to_all_internet = [
    {
      description = "allowing all egress traffic"
      from_port   = 0
      to_port     = 0
      protocol    = "-1"
      cidr_blocks = ["0.0.0.0/0"]
    }
  ]
}

resource "aws_security_group" "k3s_master_sg" {
  count       = var.create_group_for_k3s_masters ? 1 : 0
  description = "k3s server nodes requirements, source: https://docs.k3s.io/installation/requirements#networking"
  name_prefix = var.tracker_id

  ingress {
    description = "k3s HA masters: etcd communication, source: https://docs.k3s.io/installation/requirements#networking"
    from_port   = 2379
    protocol    = "tcp"
    to_port     = 2379
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "k3s HA masters: etcd communication, source: https://docs.k3s.io/installation/requirements#networking"
    from_port   = 2380
    protocol    = "tcp"
    to_port     = 2380
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "api server communication, source: https://docs.k3s.io/installation/requirements#networking"
    from_port   = 6443
    protocol    = "tcp"
    to_port     = 6443
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "k3s masters: flannel wireguard_native communication, source: https://docs.k3s.io/installation/requirements#networking"
    from_port   = 51820
    protocol    = "udp"
    to_port     = 51820
    cidr_blocks = ["0.0.0.0/0"]
  }

  dynamic "egress" {
    for_each = {for k, v in local.outgoing_to_all_internet : k => v}
    content {
      description = egress.value.description
      from_port   = egress.value.from_port
      protocol    = egress.value.protocol
      to_port     = egress.value.to_port
      cidr_blocks = egress.value.cidr_blocks
    }
  }

  dynamic "ingress" {
    for_each = {for k, v in local.incoming_ssh : k => v}
    content {
      description = ingress.value.description
      from_port   = ingress.value.from_port
      protocol    = ingress.value.protocol
      to_port     = ingress.value.from_port
      cidr_blocks = ingress.value.cidr_blocks
    }
  }

  dynamic "ingress" {
    for_each = {for k, v in local.incoming_metrics_server : k => v if var.allow_metrics_server_on_master}
    content {
      description = ingress.value.description
      from_port   = ingress.value.from_port
      protocol    = ingress.value.protocol
      to_port     = ingress.value.to_port
      cidr_blocks = ingress.value.cidr_blocks
    }
  }

  dynamic "ingress" {
    for_each = {for k, v in local.incoming_http_traffic : k => v if var.allow_incoming_http_traffic_on_master}
    content {
      description = ingress.value.description
      from_port   = ingress.value.from_port
      protocol    = ingress.value.protocol
      to_port     = ingress.value.to_port
      cidr_blocks = ingress.value.cidr_blocks
    }
  }

  dynamic "ingress" {
    for_each = {for k, v in local.k8s_node_ports : k => v if var.expose_k8s_node_ports_on_master}
    content {
      description = ingress.value.description
      from_port   = ingress.value.from_port
      protocol    = ingress.value.protocol
      to_port     = ingress.value.to_port
      cidr_blocks = ingress.value.cidr_blocks
    }
  }

  tags = {
    TrackerId = var.tracker_id
    Terraform = true
  }
}

resource "aws_security_group" "k3s_agent_sg" {
  count       = var.create_group_for_k3s_workers ? 1 : 0
  description = "k3s agent nodes, security group"
  name_prefix = var.tracker_id

  tags = {
    TrackerId = var.tracker_id
    Terraform = true
  }

  dynamic "ingress" {
    for_each = {for k, v in local.incoming_ssh : k => v}
    content {
      description = ingress.value.description
      from_port   = ingress.value.from_port
      protocol    = ingress.value.protocol
      to_port     = ingress.value.from_port
      cidr_blocks = ingress.value.cidr_blocks
    }
  }

  dynamic "ingress" {
    for_each = {for k, v in local.incoming_metrics_server : k => v if var.allow_metrics_server_on_agent}
    content {
      description = ingress.value.description
      from_port   = ingress.value.from_port
      protocol    = ingress.value.protocol
      to_port     = ingress.value.to_port
      cidr_blocks = ingress.value.cidr_blocks
    }
  }

  dynamic "ingress" {
    for_each = {for k, v in local.k8s_node_ports : k => v if var.expose_k8s_node_ports_on_agent}
    content {
      description = ingress.value.description
      from_port   = ingress.value.from_port
      protocol    = ingress.value.protocol
      to_port     = ingress.value.to_port
      cidr_blocks = ingress.value.cidr_blocks
    }
  }

  dynamic "egress" {
    for_each = {for k, v in local.outgoing_to_all_internet : k => v if var.allow_outgoing_to_all_internet_on_agent}
    content {
      description = egress.value.description
      from_port   = egress.value.from_port
      protocol    = egress.value.protocol
      to_port     = egress.value.to_port
      cidr_blocks = egress.value.cidr_blocks
    }
  }
}
