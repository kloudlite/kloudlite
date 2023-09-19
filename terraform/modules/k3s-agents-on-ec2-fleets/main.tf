resource "tls_private_key" "ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "spot_instances_ssh_key" {
  key_name   = "iac-spot"
  public_key = tls_private_key.ssh_key.public_key_openssh
}

resource "null_resource" "save_ssh_key" {
  count = var.save_ssh_key.enabled ? 1 : 0

  provisioner "local-exec" {
    command = "echo '${tls_private_key.ssh_key.private_key_pem}' > ${var.save_ssh_key.path} && chmod 600 ${var.save_ssh_key.path}"
  }
}

resource "aws_launch_template" "spot_templates" {
  for_each = {for idx, config in var.spot_nodes : idx => config}
  name     = each.key
  image_id = var.aws_ami

  key_name = aws_key_pair.spot_instances_ssh_key.key_name

  iam_instance_profile {
    name = each.value.iam_instance_profile != "" ? each.value.iam_instance_profile : null
  }

  user_data = base64encode(templatefile("${path.module}/user_data.tpl.sh", {
    k3s_server_host = var.k3s_server_dns_hostname
    k3s_token       = var.k3s_token
    node_labels     = each.value.node_labels
    node_name       = each.key
    disable_ssh     = var.disable_ssh
  }))

  block_device_mappings {
    device_name = "/dev/sda1"
    ebs {
      volume_type = each.value.root_volume_type
      volume_size = each.value.root_volume_size
    }
  }

  network_interfaces {
    #  associate_public_ip_address = each.value.allow_public_ip
    security_groups = each.value.security_groups
  }

  tag_specifications {
    resource_type = "instance"
    tags          = {
      Terraform     = "true"
      AttachesToK3s = var.k3s_server_dns_hostname
      NodeName      = each.key
    }
  }
}

data "aws_caller_identity" "current" {}

resource "aws_spot_fleet_request" "spot_fleets" {
  for_each       = {for idx, config in var.spot_nodes : idx => config}
  #   iam_fleet_role = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-ec2-spot-fleet-tagging-role"
  iam_fleet_role = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.spot_fleet_tagging_role_name}"

  instance_pools_to_use_count = 1

  target_capacity     = 1
  allocation_strategy = "priceCapacityOptimized"

  lifecycle {
    ignore_changes = [instance_pools_to_use_count]
  }

  launch_specification {
    ami           = var.aws_ami
    instance_type = each.value.instance_type
    #    subnet_id     = var.subnet_ids[0]

    #    vpc_security_group_ids = [
    #      aws_security_group.main.id,
    #    ]

    weighted_capacity = 4
    #    tags                 = local.spot_fleet_tags
    tags              = {
      Terraform     = "true"
      AttachesToK3s = var.k3s_server_dns_hostname
      NodeName      = each.key
    }
    iam_instance_profile = each.value.iam_instance_profile != "" ? each.value.iam_instance_profile : null
    user_data            = templatefile("${path.module}/user_data.tpl.sh", {
      k3s_server_host = var.k3s_server_dns_hostname
      k3s_token       = var.k3s_token
      node_labels     = each.value.node_labels
      node_name       = each.key
      disable_ssh     = var.disable_ssh
    })
  }

  #  launch_template_config {
  #    launch_template_specification {
  #      id      = aws_launch_template.spot_templates[each.key].id
  #      version = aws_launch_template.spot_templates[each.key].latest_version
  #    }
  #
  #    overrides {
  #      availability_zone = each.value.az
  #      instance_type     = each.value.instance_type
  #    }
  #  }

  fleet_type                    = "maintain"
  replace_unhealthy_instances   = true
  terminate_instances_on_delete = true
}