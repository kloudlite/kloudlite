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

resource "random_password" "k3s_token" {
  length  = 64
  special = false
}

resource "aws_security_group" "sg" {
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
  ssh_private_key = base64decode(var.ssh_private_key)
  ssh_public_key  = base64decode(var.ssh_public_key)
  master_config = [
    for i in range(1, var.master_nodes_config["count"] + 1) : {
      instance_name     = "${var.master_nodes_config["name"]}-${i}"
      instance_type     = var.master_nodes_config["instance_type"]
      ami               = var.master_nodes_config["ami"]
      availability_zone = var.master_nodes_config.availability_zones[(tonumber(i) - 1) % length(var.master_nodes_config.availability_zones)]
    }
  ]

  worker_config = [
    for i in range(1, var.worker_nodes_config["count"] + 1) : {
      instance_name     = "${var.worker_nodes_config["name"]}-${i}"
      instance_type     = var.worker_nodes_config["instance_type"]
      ami               = var.worker_nodes_config["ami"]
      availability_zone = var.master_nodes_config.availability_zones[(tonumber(i) - 1) % length(var.master_nodes_config.availability_zones)]
    }
  ]
}

resource "aws_key_pair" "k8s_masters_ssh_key" {
  key_name   = "iac-production"
  public_key = local.ssh_public_key
}

resource "aws_instance" "k8s_first_master" {
  ami               = local.master_config[0].ami
  instance_type     = local.master_config[0].instance_type
  security_groups   = [aws_security_group.sg.name]
  key_name          = aws_key_pair.k8s_masters_ssh_key.key_name
  availability_zone = local.master_config[0].availability_zone

  tags = {
    Name = local.master_config[0].instance_name
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
    private_key = local.ssh_private_key
  }

  provisioner "remote-exec" {
    inline = [
      <<-EOT
      cat > runner-config.yml <<EOF2
      runAs: primaryMaster
      primaryMaster:
        publicIP: ${self.public_ip}
        token: ${random_password.k3s_token.result}
        nodeName: ${local.master_config[0].instance_name}
        SANs:
          - ${var.cloudflare_domain}
      EOF2
      sudo ln -sf $PWD/runner-config.yml /runner-config.yml
      EOT
    ]
  }
}

resource "ssh_resource" "grab_k8s_config" {
  host        = aws_instance.k8s_first_master.public_ip
  user        = "ubuntu"
  private_key = local.ssh_private_key

  file {
    source      = "./scripts/k8s-user-account.sh"
    destination = "./k8s-user-account.sh"
    permissions = 0755
  }

  commands = [
    <<EOC
      chmod +x ./k8s-user-account.sh
      export KUBECTL='sudo k3s kubectl'

      while true; do
        if [ ! -f /etc/rancher/k3s/k3s.yaml ]; then
          # echo 'k3s yaml not found, re-checking in 1s' >> /dev/stderr
          sleep 1
          continue
        fi

        # echo "/etc/rancher/k3s/k3s.yaml file found" >> /dev/stderr

        # echo "checking whether k3s server is accepting connections" >> /dev/stderr

        lines=$(sudo k3s kubectl get nodes | wc -l)

        if [ "$lines" -lt 2 ]; then
          # echo "k3s server is not accepting connections yet, retrying in 1s ..." >> /dev/stderr
          sleep 1
          continue
        fi
        # echo "successful, k3s server is now accepting connections"
        break
      done
      ./k8s-user-account.sh >> /dev/stderr

      kubeconfig=$(cat kubeconfig.yml | sed "s|https://127.0.0.1:6443|https://${var.cloudflare_domain}:6443|" | base64 | tr -d '\n')

      echo $kubeconfig
    EOC
  ]
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
}

resource "ssh_resource" "k8s_masters" {
  for_each    = { for idx, config in local.master_config : idx => config if idx >= 1 }
  host        = aws_instance.k8s_masters[tonumber(each.key)].public_ip
  user        = "ubuntu"
  private_key = local.ssh_private_key

  when = "create"

  commands = [
    <<EOC
   cat > runner-config.yml <<EOF2
   runAs: secondaryMaster
   secondaryMaster:
     publicIP: ${aws_instance.k8s_masters[tonumber(each.key)].public_ip}
     serverIP: ${aws_instance.k8s_first_master.public_ip}
     token: ${random_password.k3s_token.result}
     nodeName: ${each.value.instance_name}
     SANs:
       - ${var.cloudflare_domain}
   EOF2

   sudo ln -sf $PWD/runner-config.yml /runner-config.yml
   sudo systemctl disable sshd.service
   sudo systemctl stop sshd.service
   sudo rm -f ~/.ssh/authorized_keys
   EOC
  ]
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
}

resource "ssh_resource" "k8s_workers" {
  for_each    = { for idx, config in local.worker_config : idx => config if idx >= 1 }
  host        = aws_instance.k8s_workers[tonumber(each.key)].public_ip
  user        = "ubuntu"
  private_key = local.ssh_private_key

  commands = [
    <<EOC
     cat > runner-config.yml <<EOF2
     runAs: agent
     agent:
       publicIP: ${aws_instance.k8s_workers[tonumber(each.key)].public_ip}
       serverIP: ${var.cloudflare_domain}
       token: ${random_password.k3s_token.result}
       nodeName: ${each.value.instance_name}
     EOF2

     sudo ln -sf $PWD/runner-config.yml /runner-config.yml
     sudo systemctl disable sshd.service
     sudo systemctl stop sshd.service
     sudo rm -f ~/.ssh/authorized_keys
    EOC
  ]
}

output "k8s_masters_public_ips" {
  value = concat([aws_instance.k8s_first_master.public_ip], [for instance in aws_instance.k8s_masters : instance.public_ip])
}

output "kubeconfig" {
  value = ssh_resource.grab_k8s_config.result
}
