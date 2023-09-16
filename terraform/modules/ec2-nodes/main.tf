resource "tls_private_key" "ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "k3s_nodes_ssh_key" {
  key_name   = "iac"
  public_key = tls_private_key.ssh_key.public_key_openssh
}

resource "null_resource" "save_ssh_key" {
  count = var.save_ssh_key.enabled ? 1 : 0

  provisioner "local-exec" {
    command = "echo '${tls_private_key.ssh_key.private_key_pem}' > ${var.save_ssh_key.path} && chmod 600 ${var.save_ssh_key.path}"
  }
}

resource "aws_security_group" "sg" {
  description = "terraform: kloudlite iac module: ec2-nodes"

  ingress {
    from_port   = 22
    protocol    = "tcp"
    to_port     = 22
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 2379
    protocol    = "tcp"
    to_port     = 2379
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 2380
    protocol    = "tcp"
    to_port     = 2380
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 6443
    protocol    = "tcp"
    to_port     = 6443
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8472
    protocol    = "udp"
    to_port     = 8472
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 9100
    protocol    = "tcp"
    to_port     = 9100
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 51820
    protocol    = "udp"
    to_port     = 51820
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 51821
    protocol    = "udp"
    to_port     = 51821
    cidr_blocks = ["0.0.0.0/0"]
  }


  ingress {
    from_port   = 10250
    protocol    = "tcp"
    to_port     = 10250
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    protocol    = "tcp"
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    protocol    = "tcp"
    to_port     = 443
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 30000
    protocol    = "tcp"
    to_port     = 32768
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 30000
    protocol    = "udp"
    to_port     = 32768
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  # lifecycle {
  #   create_before_destroy = true
  # }
}

resource "aws_iam_instance_profile" "full_s3_and_block_storage" {
  role = "EC2StorageAccess"
}

resource "aws_instance" "ec2_instances" {
  for_each          = {for idx, config in var.nodes_config : idx => config}
  ami               = var.ami
  instance_type     = each.value.instance_type
  security_groups   = each.value.security_groups
  key_name          = aws_key_pair.k3s_nodes_ssh_key.key_name
  availability_zone = each.value.az

  #  iam_instance_profile = aws_iam_instance_profile.full_s3_and_block_storage.name
  iam_instance_profile = each.value.iam_instance_profile

  tags = {
    Name      = each.key
    Terraform = true
  }

  root_block_device {
    volume_size = each.value.root_volume_size
    volume_type = each.value.root_volume_type
    encrypted   = false
    # kms_key_id  = data.aws_kms_key.customer_master_key.arn
  }
}

locals {
  nodes_with_elastic_ips = {
    for node_name, node_cfg in var.nodes_config : node_name => node_cfg
    if node_cfg.with_elastic_ip == true
  }
}

resource "aws_eip" "elastic_ips" {
  for_each = local.nodes_with_elastic_ips
}

resource "aws_eip_association" "k3s_masters_elastic_ips_association" {
  for_each      = local.nodes_with_elastic_ips
  instance_id   = aws_instance.ec2_instances[each.key].id
  allocation_id = aws_eip.elastic_ips[each.key].id
}
