module "cluster" {
  source                           = "../../../terraform/bundles/aws/cluster/"
  cluster_name                     = var.cluster_name
  aws_region                       = var.aws_region
  vpc_id                           = var.vpc_id
  cluster_state                    = var.cluster_state
  master_nodes                     = var.master_nodes
  kloudlite_release                = var.kloudlite_release
  base_domain                      = var.base_domain
  master_node_iam_instance_profile = var.master_node_iam_instance_profile
}
