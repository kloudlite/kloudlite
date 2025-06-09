module "kl-work-machine-on-aws" {
  source               = "../../../terraform/bundles/aws/work-machine"
  name                 = var.name
  vpc_id               = var.vpc_id
  trace_id             = var.trace_id
  k3s_server_host      = var.k3s_server_host
  k3s_agent_token      = var.k3s_agent_token
  k3s_version          = var.k3s_version
  ami                  = var.ami
  instance_type        = var.instance_type
  instance_state       = var.instance_state
  availability_zone    = var.availability_zone
  iam_instance_profile = var.iam_instance_profile
  root_volume_size     = var.root_volume_size
  root_volume_type     = var.root_volume_type
  security_group_ids   = var.security_group_ids
  subnet_id            = var.subnet_id
}
