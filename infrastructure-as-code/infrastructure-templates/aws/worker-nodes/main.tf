module "kl-worker-nodes-on-aws" {
  source     = "../../../terraform/bundles/aws/worker-nodes"
  aws_region = var.aws_region

  k3s_join_token             = var.k3s_join_token
  k3s_server_public_dns_host = var.k3s_server_public_dns_host

  tracker_id           = "${var.tracker_id}-workers"
  extra_agent_args     = var.extra_agent_args
  save_ssh_key_to_path = var.save_ssh_key_to_path
  tags                 = var.tags
  kloudlite_release    = var.kloudlite_release

  vpc_id                        = var.vpc_id
  availability_zone             = var.availability_zone
  ec2_nodepool                  = var.ec2_nodepool
  nodepool_name                 = var.nodepool_name
  spot_nodepool                 = var.spot_nodepool
  vpc_subnet_id                 = var.vpc_subnet_id
  k3s_download_url              = var.k3s_download_url
  kloudlite_runner_download_url = var.kloudlite_runner_download_url
}
