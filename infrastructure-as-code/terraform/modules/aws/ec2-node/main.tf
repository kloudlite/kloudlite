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

locals {
  resource_tracker_id = "${var.tracker_id}-${var.node_name}"
  #  k3s_data_volume_device = "/dev/sdf"
}

#resource "aws_ebs_volume" "k3s_data_volume" {
#  availability_zone = var.availability_zone
#  size              = 10  # Size in GiB
#  type              = "gp3"  # Adjust type based on needs
#  tags              = {
#    Name         = "${local.resource_tracker_id}-k3s-data-volume"
#    TrackerId    = var.tracker_id
#    KloudliteIAC = true
#  }
#}

resource "aws_instance" "ec2_instance" {
  ami           = var.ami
  instance_type = var.instance_type

  security_groups        = var.vpc == null ? var.security_groups : null
  vpc_security_group_ids = var.vpc != null ? var.vpc.vpc_security_group_ids : null
  subnet_id              = var.vpc != null ? var.vpc.subnet_id : null

  key_name          = var.ssh_key_name
  availability_zone = var.availability_zone

  lifecycle {
    replace_triggered_by = [null_resource.lifecycle_resource]
    ignore_changes       = [ami, instance_type, user_data_base64]
  }

  depends_on = [null_resource.variable_validation]

  user_data_base64     = var.user_data_base64 != "" ? var.user_data_base64 : null
  iam_instance_profile = var.iam_instance_profile != "" ? var.iam_instance_profile : null

  tags = merge({
    Name         = local.resource_tracker_id
    KloudliteIAC = true
  }, var.tags)

  root_block_device {
    volume_size = var.root_volume_size
    volume_type = var.root_volume_type
    encrypted   = false
    # kms_key_id  = data.aws_kms_key.customer_master_key.arn
  }
}

#resource "aws_volume_attachment" "ebs_attach" {
#  device_name = local.k3s_data_volume_device
#  volume_id   = aws_ebs_volume.k3s_data_volume.id
#  instance_id = aws_instance.ec2_instance.id
#}
