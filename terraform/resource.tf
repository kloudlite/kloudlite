terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }

    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "4.13.0"
    }
  }

  required_version = ">= 1.2.0"
}

provider "aws" {
  region     = var.region
  access_key = var.access_key
  secret_key = var.secret_key
}

data "aws_ami" "latest-ubuntu" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_security_group" "sg" {
  # name = "${var.node_name}-sg"

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

locals {
  master_azs = split(",", var.master_nodes_config["availability_zones"])
  master_config = [
    for i in range(1, var.master_nodes_config["count"] + 1) : {
      instance_name     = "${var.master_nodes_config["name"]}-${i}"
      instance_type     = var.master_nodes_config["instance_type"]
      ami               = var.master_nodes_config["ami"]
      availability_zone = local.master_azs[(tonumber(i) - 1) % length(local.master_azs)]
      # availability_zone = local.master_azs_map[(tonumber(i) - 1) % length(local.master_azs)].zone
    }
  ]

  worker_azs = split(",", var.worker_nodes_config["availability_zones"])
  worker_config = [
    for i in range(1, var.worker_nodes_config["count"] + 1) : {
      instance_name     = "${var.worker_nodes_config["name"]}-${i}"
      instance_type     = var.worker_nodes_config["instance_type"]
      ami               = var.worker_nodes_config["ami"]
      availability_zone = local.worker_azs[(tonumber(i) - 1) % length(local.worker_azs)]
      # availability_zone = local.worker_azs_map[(i-1) % length(local.worker_azs)].zone
    }
  ]
}

resource "aws_key_pair" "k8s_masters_ssh_key" {
  key_name   = "iac-production"
  public_key = file(var.public_key_path)
}

# resource "aws_eip" "k8s_first_master_public_ip" {
#   instance = aws_instance.k8s_first_master.id
# }

resource "aws_instance" "k8s_first_master" {
  #  ami               = var.master_nodes_config["ami"]
  ami               = local.master_config[0].ami
  instance_type     = local.master_config[0].instance_type
  security_groups   = [aws_security_group.sg.name]
  key_name          = aws_key_pair.k8s_masters_ssh_key.key_name
  availability_zone = local.master_config[0].availability_zone

  tags = {
    #    Name = "${var.master_nodes_config["name"]}-1"
    Name = local.master_config[0].instance_name
  }

  root_block_device {
    volume_size = 100 # in GB <<----- I increased this!
    volume_type = "standard"
    encrypted   = false
    # kms_key_id  = data.aws_kms_key.customer_master_key.arn
  }

  connection {
    type        = "ssh"
    user        = "ubuntu"
    host        = self.public_ip
    private_key = file(var.private_key_path)
  }

  provisioner "remote-exec" {
    inline = [
      <<-EOT
      cat > runner-config.yml <<EOF2
      runAs: primaryMaster
      primaryMaster:
        publicIP: ${self.public_ip}
        token: ${var.k3s_token}
        nodeName: ${local.master_config[0].instance_name}
        SANs:
          - ${var.cloudflare.domain}
      EOF2
      sudo ln -sf $PWD/runner-config.yml /runner-config.yml
      EOT
    ]
  }
}

resource "aws_instance" "k8s_masters" {
  for_each          = { for idx, config in local.master_config : idx => config if idx >= 1 }
  ami               = var.master_nodes_config["ami"]
  instance_type     = each.value.instance_type
  security_groups   = [aws_security_group.sg.name]
  key_name          = aws_key_pair.k8s_masters_ssh_key.key_name
  availability_zone = each.value.availability_zone

  tags = {
    Name = each.value.instance_name
  }

  root_block_device {
    volume_size = 100 # in GB <<----- I increased this!
    volume_type = "standard"
    encrypted   = false
    # kms_key_id  = data.aws_kms_key.customer_master_key.arn
  }

  connection {
    type        = "ssh"
    user        = "ubuntu"
    host        = self.public_ip
    private_key = file(var.private_key_path)
  }

  provisioner "remote-exec" {
    inline = [
      <<-EOT
      cat > runner-config.yml <<EOF2
      runAs: secondaryMaster
      secondaryMaster:
        publicIP: ${self.public_ip}
        serverIP: ${aws_instance.k8s_first_master.public_ip}
        token: ${var.k3s_token}
        nodeName: ${each.value.instance_name}
        SANs:
          - ${var.cloudflare.domain}
      EOF2
      sudo ln -sf $PWD/runner-config.yml /runner-config.yml
      # sudo rm -f ~/.ssh/authorized_keys
      EOT
    ]
  }
}


resource "aws_instance" "k8s_workers" {
  for_each          = { for idx, config in local.worker_config : idx => config }
  ami               = var.worker_nodes_config["ami"]
  instance_type     = each.value.instance_type
  security_groups   = [aws_security_group.sg.name]
  availability_zone = each.value.availability_zone
  key_name          = aws_key_pair.k8s_masters_ssh_key.key_name

  tags = {
    Name = each.value.instance_name
  }

  root_block_device {
    volume_size = 100
    volume_type = "standard"
    encrypted   = false
    # kms_key_id  = data.aws_kms_key.customer_master_key.arn
  }

  connection {
    type        = "ssh"
    user        = "ubuntu"
    host        = self.public_ip
    private_key = file(var.private_key_path)
  }

  provisioner "remote-exec" {
    inline = [
      <<-EOT
      cat > runner-config.yml <<EOF2
      runAs: agent
      agent:
        publicIP: ${self.public_ip}
        # serverIP: ${aws_instance.k8s_first_master.public_ip}
        serverIP: ${var.cloudflare.domain}
        token: ${var.k3s_token}
        nodeName: ${each.value.instance_name}
      EOF2
      sudo ln -sf $PWD/runner-config.yml /runner-config.yml
      EOT
    ]
  }
}

resource "null_resource" "grab_kube_config" {
  provisioner "local-exec" {
    # command = "scp -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i ${var.private_key_path} ubuntu@${aws_instance.k8s_first_master.public_ip}:/etc/rancher/k3s/k3s.yaml kubeconfig.yaml"
    command = "ssh -i ${var.private_key_path} ubuntu@${aws_instance.k8s_first_master.public_ip} -C 'while true; do [ ! -f /etc/rancher/k3s/k3s.yaml ] && echo 'k3s yaml not found, re-checking in 1s' && sleep 1 && continue;  sudo cat /etc/rancher/k3s/k3s.yaml; break; done' > kubeconfig.yaml"
    # command = "ssh -i ${var.private_key_path} ubuntu@${aws_instance.k8s_first_master.public_ip} -C 'sudo cat /etc/rancher/k3s/k3s.yaml' > kubeconfig.yaml"
  }
}

output "k8s_masters_public_ips" {
  value = concat([aws_instance.k8s_first_master.public_ip], [for instance in aws_instance.k8s_masters : instance.public_ip])
}

# output "node-ip" {
#   value = aws_instance.k8s_masters[*].public_ip
# }
#
# output "node-name" {
#   value = var.node_name
# }
