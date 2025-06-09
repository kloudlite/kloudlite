module "ssh-rsa-key" {
  source = "../../../modules/common/ssh-rsa-key"
}

resource "aws_key_pair" "ssh_key_pair" {
  key_name   = "${var.trace_id}-ssh-key-pair"
  public_key = module.ssh-rsa-key.public_key
}

resource "aws_ebs_volume" "external_volume" {
  availability_zone = var.availability_zone
  size              = 100
  type              = "gp3"
  encrypted         = true

  tags = {
    Name = "${var.trace_id}-work-machine-volume"
  }
}

module "ec2-node" {
  source   = "../../../modules/aws/ec2-node-v2"
  trace_id = var.trace_id

  name = var.name

  ami    = var.ami
  vpc_id = var.vpc_id

  instance_type  = var.instance_type
  instance_state = var.instance_state

  availability_zone    = var.availability_zone
  iam_instance_profile = var.iam_instance_profile
  root_volume_size     = var.root_volume_size
  root_volume_type     = var.root_volume_type
  ssh_key_name         = aws_key_pair.ssh_key_pair.key_name

  user_data_base64 = base64encode(
    templatefile("${path.module}/launch-script.sh", {
      k3s_server_host = var.k3s_server_host
      k3s_agent_token = var.k3s_agent_token
      k3s_version     = var.k3s_version
      node_name       = var.name,
  }))

  security_group_ids = var.security_group_ids
  subnet_id          = var.subnet_id
}

resource "aws_volume_attachment" "volume_attachment" {
  instance_id = module.ec2-node.instance_id
  volume_id   = aws_ebs_volume.external_volume.id
  device_name = "/dev/xvdb"
}
