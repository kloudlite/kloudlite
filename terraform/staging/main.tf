module "k3s-on-ec2" {
  source = "../modules/k3s-HA-on-ec2"

  master_nodes_config = {}
  worker_nodes_config = {}
  ssh_private_key = ""
  ssh_public_key = ""
  domain = ""
}
