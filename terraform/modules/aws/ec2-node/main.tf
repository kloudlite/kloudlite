locals {
  check_satisifies_minimum_root_volume_size = {
    error_message = "when node is nvidia gpu enabled, root volume size must be greater than 75GiB, otherwise greater than 50Gi"
    condition     = var.root_volume_size >= (var.is_nvidia_gpu_node == true  ? 75 : 50)
  }
}

resource "null_resource" "variable_validation" {
  lifecycle {
    precondition {
      error_message = local.check_satisifies_minimum_root_volume_size.error_message
      condition     = local.check_satisifies_minimum_root_volume_size.condition
    }
  }
}

resource "null_resource" "lifecycle_resource" {
  depends_on = [null_resource.variable_validation]
  triggers   = {
    on_recreate = var.last_recreated_at
  }
}

resource "aws_instance" "ec2_instance" {
  ami               = var.ami
  instance_type     = var.instance_type
  security_groups   = var.security_groups
  key_name          = var.ssh_key_name
  availability_zone = var.availability_zone

  lifecycle {
    replace_triggered_by = [null_resource.lifecycle_resource]
  }

  depends_on = [null_resource.variable_validation]

  user_data_base64 = var.user_data_base64 != "" ? var.user_data_base64 : null

  iam_instance_profile = var.iam_instance_profile != "" ? var.iam_instance_profile : null

  tags = {
    Name      = "${var.tracker_id}-${var.node_name}"
    Terraform = true
  }

  root_block_device {
    volume_size = var.root_volume_size
    volume_type = var.root_volume_type
    encrypted   = false
    # kms_key_id  = data.aws_kms_key.customer_master_key.arn
  }
}

#locals {
#  nodes_with_elastic_ips = {
#    for node_name, node_cfg in var.nodes_config : node_name => node_cfg
#    if node_cfg.with_elastic_ip == true
#  }
#}

#resource "aws_eip" "elastic_ips" {
#  for_each   = local.nodes_with_elastic_ips
#  depends_on = [aws_instance.ec2_instances]
#  tags       = {
#    Name      = "${each.key}-elastic-ip"
#    Terraform = true
#  }
#}
#
#resource "aws_eip_association" "k3s_masters_elastic_ips_association" {
#  for_each      = local.nodes_with_elastic_ips
#  depends_on    = [aws_eip.elastic_ips]
#  instance_id   = aws_instance.ec2_instances[each.key].id
#  allocation_id = aws_eip.elastic_ips[each.key].id
#}
