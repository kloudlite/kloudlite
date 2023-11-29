module "ec2-nodes" {
  source               = "../ec2-node"
  for_each             = {for name, cfg in var.nodes : name => cfg}
  tracker_id           = var.tracker_id
  ami                  = var.ami
  availability_zone    = var.availability_zone
  iam_instance_profile = var.iam_instance_profile
  instance_type        = var.instance_type
  is_nvidia_gpu_node   = var.nvidia_gpu_enabled
  node_name            = each.key
  root_volume_size     = var.root_volume_size
  root_volume_type     = var.root_volume_type
  security_groups      = var.security_groups
  last_recreated_at    = each.value.last_recreated_at
  ssh_key_name         = var.ssh_key_name
  user_data_base64     = each.value.user_data_base64 != null ? each.value.user_data_base64 : ""
}