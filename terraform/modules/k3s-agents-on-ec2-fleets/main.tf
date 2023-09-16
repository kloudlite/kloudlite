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

resource "aws_launch_template" "spot-templates" {
  for_each = {for idx, config in var.spot_nodes : idx => config}
  name     = each.key
  image_id = var.aws_ami

  key_name = aws_key_pair.spot_instances_ssh_key.key_name

  iam_instance_profile {
    name = each.value.iam_instance_profile
  }

  user_data = base64encode(templatefile("${path.module}/user_data.tpl.sh", {
    k3s_server_host = var.k3s_server_host
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
    associate_public_ip_address = true
    security_groups             = each.value.security_groups
  }

  tag_specifications {
    resource_type = "instance"
    tags          = {
      Terraform     = "true"
      AttachesToK3s = var.k3s_server_host
    }
  }
}

data "aws_caller_identity" "current" {}

#resource "aws_ec2_fleet" "spot-fleets" {
#  for_each = {for idx, config in var.spot_nodes : idx => config}
#  launch_template_config {
#    launch_template_specification {
#      launch_template_id = aws_launch_template.spot-templates[each.key].id
#      version            = aws_launch_template.spot-templates[each.key].latest_version
#    }
#
#    override {
#      availability_zone = each.value.az
#      instance_type     = each.value.instance_type
#      #      max_price     = each.value.max_price_per_hour
#    }
#  }
#
#  spot_options {
#    allocation_strategy = "lowestPrice"
#  }
#
#  target_capacity_specification {
#    default_target_capacity_type = "spot"
#    spot_target_capacity         = 1
#    total_target_capacity        = 1
#  }
#
#  type                                = "maintain"
#  replace_unhealthy_instances         = true
#  terminate_instances_with_expiration = true
#}

resource "aws_spot_fleet_request" "spot-fleets" {
  for_each       = {for idx, config in var.spot_nodes : idx => config}
  iam_fleet_role = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-ec2-spot-fleet-tagging-role"

  instance_pools_to_use_count = 1

  target_capacity               = 1
  on_demand_target_capacity     = 0
  allocation_strategy           = "priceCapacityOptimized"
  on_demand_allocation_strategy = "lowestPrice"

  launch_template_config {
    launch_template_specification {
      id      = aws_launch_template.spot-templates[each.key].id
      version = aws_launch_template.spot-templates[each.key].latest_version
    }

    overrides {
      availability_zone = each.value.az
      instance_type     = each.value.instance_type
    }
  }

  fleet_type                    = "maintain"
  replace_unhealthy_instances   = true
  terminate_instances_on_delete = true
}