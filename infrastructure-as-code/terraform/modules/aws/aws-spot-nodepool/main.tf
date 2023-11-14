module "spot-fleet-request" {
  source                       = "../spot-fleet-request"
  for_each                     = {for name, cfg in var.nodes : name => cfg}
  ami                          = var.ami
  availability_zone            = var.availability_zone
  cpu_node                     = var.cpu_node
  gpu_node                     = var.gpu_node
  iam_instance_profile         = var.iam_instance_profile
  node_name                    = each.key
  tracker_id                   = var.tracker_id
  root_volume_size             = var.root_volume_size
  root_volume_type             = var.root_volume_type
  security_groups              = var.security_groups
  spot_fleet_tagging_role_name = var.spot_fleet_tagging_role_name
  ssh_key_name                 = var.ssh_key_name
  user_data_base64             = each.value.user_data_base64
  last_recreated_at            = each.value.last_recreated_at != null ? each.value.last_recreated_at : 0
}
