resource "tls_private_key" "ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "k3s_nodes_ssh_key" {
  key_name   = "iac"
  public_key = tls_private_key.ssh_key.public_key_openssh
}

resource "null_resource" "save_ssh_key_if_ssh_enabled" {
  count = (!var.disable_ssh && length(var.save_ssh_key_as) == true) ? 1 : 0
  triggers = {
    ssh_key = tls_private_key.ssh_key.private_key_pem
  }

  provisioner "local-exec" {
    command = "mkdir -p $(dirname ${var.save_ssh_key_as}) && echo '${tls_private_key.ssh_key.private_key_pem}' > ${var.save_ssh_key_as} && chmod 600 ${var.save_ssh_key_as}"
  }
}

locals {
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

resource "random_password" "k3s_token" {
  length  = 64
  special = false
}

resource "aws_instance" "k3s_primary_master" {
  ami               = local.master_config[0].ami
  instance_type     = local.master_config[0].instance_type
  security_groups   = [aws_security_group.sg.name]
  key_name          = aws_key_pair.k3s_nodes_ssh_key.key_name
  availability_zone = local.master_config[0].availability_zone

  tags = {
    Name      = local.master_config[0].instance_name
    Terraform = true
  }

  root_block_device {
    volume_size = 100
    volume_type = "standard"
    encrypted   = false
    # kms_key_id  = data.aws_kms_key.customer_master_key.arn
  }
}

resource "null_resource" "setup_k3s_on_primary_master" {
  connection {
    type        = "ssh"
    user        = "ubuntu"
    host        = aws_instance.k3s_primary_master.public_ip
    private_key = tls_private_key.ssh_key.private_key_pem
  }

  provisioner "remote-exec" {
    inline = [
      <<-EOT
      cat > runner-config.yml <<EOF2
      runAs: primaryMaster
      primaryMaster:
        publicIP: ${aws_instance.k3s_primary_master.public_ip}
        token: ${random_password.k3s_token.result}
        nodeName: ${local.master_config[0].instance_name}
        SANs:
          - ${var.domain}
      EOF2
      sudo ln -sf $PWD/runner-config.yml /runner-config.yml
      EOT
    ]
  }
}

resource "ssh_resource" "grab_k8s_config" {
  host        = aws_instance.k3s_primary_master.public_ip
  user        = "ubuntu"
  private_key = tls_private_key.ssh_key.private_key_pem

  file {
    source      = "${path.module}/scripts/k8s-user-account.sh"
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

    kubeconfig=$(cat kubeconfig.yml | sed "s|https://127.0.0.1:6443|https://${var.domain}:6443|" | base64 | tr -d '\n')

    printf "$kubeconfig"

    if [ "${var.disable_ssh}" == "true" ]; then
      sudo systemctl disable sshd.service
      sudo systemctl stop sshd.service
      sudo rm -f ~/.ssh/authorized_keys
    fi

    EOC
  ]
}

resource "aws_instance" "k3s_secondary_masters" {
  for_each          = { for idx, config in local.master_config : idx => config if idx >= 1 }
  ami               = var.master_nodes_config["ami"]
  instance_type     = each.value.instance_type
  security_groups   = [aws_security_group.sg.name]
  key_name          = aws_key_pair.k3s_nodes_ssh_key.key_name
  availability_zone = each.value.availability_zone

  tags = {
    Name      = each.value.instance_name
    Terraform = true
  }

  root_block_device {
    volume_size = 100 # in GB <<----- I increased this!
    volume_type = "standard"
    encrypted   = false
    # kms_key_id  = data.aws_kms_key.customer_master_key.arn
  }
}

resource "null_resource" "setup_k3s_on_secondary_masters" {
  for_each          = { for idx, config in local.master_config : idx => config if idx >= 1 }

  connection {
    type        = "ssh"
    user        = "ubuntu"
    host        = aws_instance.k3s_secondary_masters[tonumber(each.key)].public_ip
    private_key = tls_private_key.ssh_key.private_key_pem
  }

  provisioner "remote-exec" {
    inline = [
      <<-EOC
      cat > runner-config.yml <<EOF2
      runAs: secondaryMaster
      secondaryMaster:
        publicIP: ${aws_instance.k3s_secondary_masters[tonumber(each.key)].public_ip}
        serverIP: ${aws_instance.k3s_primary_master.public_ip}
        token: ${random_password.k3s_token.result}
        nodeName: ${each.value.instance_name}
        SANs:
          - ${var.domain}
      EOF2

      sudo ln -sf $PWD/runner-config.yml /runner-config.yml
      if [ "${var.disable_ssh}" == "true" ]; then
        sudo systemctl disable sshd.service
        sudo systemctl stop sshd.service
        sudo rm -f ~/.ssh/authorized_keys
      fi
      EOC
    ]
  }
}

resource "aws_instance" "k3s_agents" {
  for_each          = { for idx, config in local.worker_config : idx => config }
  ami               = var.worker_nodes_config["ami"]
  instance_type     = each.value.instance_type
  security_groups   = [aws_security_group.sg.name]
  availability_zone = each.value.availability_zone
  key_name          = aws_key_pair.k3s_nodes_ssh_key.key_name

  tags = {
    Name = each.value.instance_name
    Terraform = true
  }

  root_block_device {
    volume_size = 100
    volume_type = "standard"
    encrypted   = false
    # kms_key_id  = data.aws_kms_key.customer_master_key.arn
  }
}

resource "null_resource" "setup_k3s_on_agents" {
  for_each = { for idx, config in local.worker_config : idx => config }
  connection {
    type        = "ssh"
    user        = "ubuntu"
    host        = aws_instance.k3s_agents[tonumber(each.key)].public_ip
    private_key = tls_private_key.ssh_key.private_key_pem
  }

  provisioner "remote-exec" {
    inline = [
      <<-EOC
      cat >runner-config.yml <<EOF2
      runAs: agent
      agent:
        publicIP: ${aws_instance.k3s_agents[tonumber(each.key)].public_ip}
        serverIP: ${var.domain}
        token: ${random_password.k3s_token.result}
        nodeName: ${each.value.instance_name}
      EOF2

      sudo ln -sf $PWD/runner-config.yml /runner-config.yml
      if [ "${var.disable_ssh}" == "true" ]; then
        sudo systemctl disable sshd.service
        sudo systemctl stop sshd.service
        sudo rm -f ~/.ssh/authorized_keys
      fi
      EOC
    ]
  }
}

