resource "tls_private_key" "ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "random_id" "id" {
  byte_length = 30
}

resource "aws_key_pair" "spot_instances_ssh_key" {
  key_name   = "iac-${random_id.id.hex}-spot"
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
  name     = "iac-${random_id.id.hex}-${each.key}"
  image_id = var.aws_ami

  key_name = aws_key_pair.spot_instances_ssh_key.key_name

  iam_instance_profile {
    name = each.value.iam_instance_profile != "" ? each.value.iam_instance_profile : null
  }

  user_data = base64encode(templatefile("${path.module}/user_data.tpl.sh", {
    k3s_server_host     = var.k3s_server_dns_hostname
    k3s_token           = var.k3s_token
    node_labels         = each.value.node_labels
    node_name           = each.key
    disable_ssh         = var.disable_ssh
    #    is_nvidia_gpu_node = each.value.is_nvidia_gpu_node
    is_nvidia_gpu_node  = false
    nvidia_gpu_template = file("${path.module}/../scripts/nvidia-gpu-post-k3s-start.sh")
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
      Name          = each.key
      Terraform     = "true"
      AttachesToK3s = var.k3s_server_dns_hostname
      NodeName      = each.key
    }
  }
}

data "aws_caller_identity" "current" {}

resource "aws_spot_fleet_request" "spot_fleets" {
  for_each = {for idx, config in var.spot_nodes : idx => config}

  iam_fleet_role = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.spot_fleet_tagging_role_name}"

  instance_pools_to_use_count   = 1
  target_capacity               = 1
  allocation_strategy           = "lowestPrice"
  on_demand_allocation_strategy = "lowestPrice"

  lifecycle {
    ignore_changes = [instance_pools_to_use_count, iam_fleet_role]
  }

  launch_template_config {
    launch_template_specification {
      id      = aws_launch_template.spot_templates[each.key].id
      version = aws_launch_template.spot_templates[each.key].latest_version
    }

    overrides {
      availability_zone = each.value.az != "" ? each.value.az : null
      instance_requirements {
        burstable_performance = "excluded"
        instance_generations  = ["current"]
        vcpu_count {
          min = each.value.vcpu.min
          max = each.value.vcpu.max
        }
        memory_gib_per_vcpu {
          min = each.value.memory_per_vcpu.min
          max = each.value.memory_per_vcpu.max
        }
        memory_mib {
          min = 1024 * each.value.memory_per_vcpu.min * each.value.vcpu.min
          max = 1024 * each.value.memory_per_vcpu.max * each.value.vcpu.max
        }
      }
    }
  }

  tags = {
    Name             = each.key
    AvailabilityZone = each.value.az
  }

  fleet_type                      = "maintain"
  replace_unhealthy_instances     = true
  terminate_instances_on_delete   = true
  instance_interruption_behaviour = "stop"
}
