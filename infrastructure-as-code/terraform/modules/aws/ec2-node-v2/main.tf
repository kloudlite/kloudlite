resource "null_resource" "variable_validation" {
  lifecycle {
    precondition {
      error_message = "when node is nvidia gpu enabled, root volume size must be greater than 75GiB, otherwise greater than 50Gi"
      condition     = var.root_volume_size >= (var.is_nvidia_gpu_node == true ? 75 : 50)
    }
  }
}

resource "aws_instance" "ec2_instance" {
  ami           = var.ami
  instance_type = var.instance_type

  vpc_security_group_ids = var.security_group_ids
  subnet_id              = var.subnet_id

  key_name          = var.ssh_key_name
  availability_zone = var.availability_zone

  depends_on = [null_resource.variable_validation]

  user_data_base64     = var.user_data_base64 != "" ? var.user_data_base64 : null
  iam_instance_profile = var.iam_instance_profile != "" ? var.iam_instance_profile : null

  tags = merge({
    Name                     = "${var.trace_id}-${var.name}"
    kloudlite_infrastructure = true
  }, var.tags)

  root_block_device {
    volume_size = var.root_volume_size
    volume_type = var.root_volume_type
    encrypted   = false
  }
}

resource "aws_ec2_instance_state" "ec2_instance" {
  instance_id = aws_instance.ec2_instance.id
  state       = var.instance_state
}
