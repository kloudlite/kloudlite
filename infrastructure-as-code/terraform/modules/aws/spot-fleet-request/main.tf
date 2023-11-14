locals {
  check_either_cpu_or_gpu_node = {
    condition     = (var.cpu_node == null && var.gpu_node != null) || (var.cpu_node != null && var.gpu_node == null)
    error_message = "a node can be either a cpu_node or a gpu_node, only one of them can be set at once"
  }

  check_satisifies_minimum_root_volume_size = {
    error_message = "when node is nvidia gpu enabled, root volume size must be greater than 75GiB, otherwise greater than 50Gi"
    condition     = var.root_volume_size >= (var.gpu_node != null ? 75 : 50)
  }
}

resource "null_resource" "variable_validation" {
  lifecycle {
    precondition {
      error_message = local.check_either_cpu_or_gpu_node.error_message
      condition     = local.check_either_cpu_or_gpu_node.condition
    }
    precondition {
      error_message = local.check_satisifies_minimum_root_volume_size.error_message
      condition     = local.check_satisifies_minimum_root_volume_size.condition
    }
  }
}

resource "aws_launch_template" "spot_template" {
  name     = "lt-${var.tracker_id}-${var.node_name}"
  image_id = var.ami

  default_version = 1

  key_name   = var.ssh_key_name
  depends_on = [null_resource.variable_validation]

  iam_instance_profile {
    name = var.iam_instance_profile != "" ? var.iam_instance_profile : null
  }

  lifecycle {
    ignore_changes = [iam_instance_profile]
  }

  user_data = var.user_data_base64

  block_device_mappings {
    device_name = "/dev/sda1"
    ebs {
      volume_type = var.root_volume_type
      volume_size = var.root_volume_size
    }
  }

  network_interfaces {
    associate_public_ip_address = true
    security_groups             = var.security_groups
  }

  tag_specifications {
    resource_type = "instance"
    tags          = {
      Name        = "${var.tracker_id}-${var.node_name}"
      Terraform   = "true"
      ReferenceId = var.tracker_id
    }
  }
}

data "aws_caller_identity" "current" {}

resource "null_resource" "lifecycle_resource" {
  depends_on = [null_resource.variable_validation]
  triggers   = {
    on_recreate = var.last_recreated_at
  }
}

resource "aws_spot_fleet_request" "cpu_spot_fleet" {
  count          = var.cpu_node != null ? 1 : 0
  iam_fleet_role = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.spot_fleet_tagging_role_name}"
  depends_on     = [null_resource.variable_validation]

  instance_pools_to_use_count   = 1
  target_capacity               = 1
  allocation_strategy           = "lowestPrice"
  on_demand_allocation_strategy = "lowestPrice"

  lifecycle {
    ignore_changes       = [instance_pools_to_use_count]
    replace_triggered_by = [null_resource.lifecycle_resource]
  }

  launch_template_config {
    launch_template_specification {
      id      = aws_launch_template.spot_template.id
      #      version = var.recreate ?  aws_launch_template.spot_template.latest_version : 1
      version = 1
    }

    overrides {
      availability_zone = var.availability_zone != "" ? var.availability_zone : null
      instance_requirements {
        burstable_performance = "excluded"
        instance_generations  = ["current"]

        vcpu_count {
          min = var.cpu_node.vcpu.min
          max = var.cpu_node.vcpu.max
        }
        memory_gib_per_vcpu {
          min = var.cpu_node.memory_per_vcpu.min
          max = var.cpu_node.memory_per_vcpu.max
        }
        memory_mib {
          min = 1024 * var.cpu_node.memory_per_vcpu.min * var.cpu_node.vcpu.min
          max = 1024 * var.cpu_node.memory_per_vcpu.max * var.cpu_node.vcpu.max
        }
      }
    }
  }

  tags = {
    Name               = "${var.tracker_id}-${var.node_name}"
    AvailabilityZone   = var.availability_zone
    IamInstanceProfile = var.iam_instance_profile
    IamFleetRole       = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.spot_fleet_tagging_role_name}"
  }

  fleet_type                      = "maintain"
  replace_unhealthy_instances     = true
  terminate_instances_on_delete   = true
  instance_interruption_behaviour = "terminate" # can be of one of "terminate", or "stop"
}

resource "aws_spot_fleet_request" "gpu_spot_fleet" {
  count          = var.gpu_node != null ? 1 : 0
  iam_fleet_role = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.spot_fleet_tagging_role_name}"
  depends_on     = [null_resource.variable_validation]

  instance_pools_to_use_count   = 1
  target_capacity               = 1
  allocation_strategy           = "lowestPrice"
  on_demand_allocation_strategy = "lowestPrice"

  lifecycle {
    ignore_changes = [instance_pools_to_use_count, iam_fleet_role]
  }

  launch_template_config {
    launch_template_specification {
      id      = aws_launch_template.spot_template.id
      #      version = aws_launch_template.spot_templates[each.key].latest_version
      #      version = var.recreate ?  aws_launch_template.spot_template.latest_version : 1
      version = 1
    }

    dynamic "overrides" {
      for_each = {for idx, config in var.gpu_node.instance_types : idx => config}
      content {
        availability_zone = var.availability_zone
        instance_type     = overrides.value
        #        spot_price        = null
        #        priority          = 0
        #        subnet_id         = null
        #        weighted_capacity = 0
      }
    }
  }

  tags = {
    Name               = "${var.tracker_id}-${var.node_name}"
    AvailabilityZone   = var.availability_zone
    IamInstanceProfile = var.iam_instance_profile
    IamFleetRole       = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${var.spot_fleet_tagging_role_name}"
  }

  fleet_type                      = "maintain"
  replace_unhealthy_instances     = true
  terminate_instances_on_delete   = true
  instance_interruption_behaviour = "terminate" # can be of one of "terminate", or "stop"
}
