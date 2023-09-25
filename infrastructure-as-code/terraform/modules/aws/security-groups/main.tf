resource "aws_security_group" "allows_ssh" {
  ingress {
    description = "required during terraform apply, to execute k3s commands, on the node"
    from_port   = 22
    protocol    = "tcp"
    to_port     = 22
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "exposes_k8s_node_ports" {
  ingress {
    description = "allowing node-port range for kubernetes services (tcp), exposed with node-port"
    from_port   = 30000
    protocol    = "tcp"
    to_port     = 32768
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "allowing node-port range for kubernetes services (udp), exposed with node-port"
    from_port   = 30000
    protocol    = "udp"
    to_port     = 32768
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "nodes_can_access_internet" {
  egress {
    description = "allowing all egress traffic from nodes"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "allow_metrics_server" {
  description = "k3s metrics server: source: https://docs.k3s.io/installation/requirements#networking"
  ingress {
    from_port   = 10250
    protocol    = "tcp"
    to_port     = 10250
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "allows_incoming_http_traffic" {
  ingress {
    description = "allows http communication from outside to the cluster"
    from_port   = 80
    protocol    = "tcp"
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "allows https communication from outside to the cluster"
    from_port   = 443
    protocol    = "tcp"
    to_port     = 443
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "k3s_server_nodes_requirements" {
  description = "k3s server nodes requirements, source: https://docs.k3s.io/installation/requirements#networking"

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

  egress {
    description = "allowing all egress traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
