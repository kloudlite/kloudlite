module "kl-master-nodes-on-aws" {
  source                    = "../../terraform/bundles/kl-master-nodes-on-aws"
  aws_region                = var.aws_region
  enable_nvidia_gpu_support = var.enable_nvidia_gpu_support
  k3s_masters               = var.k3s_masters
  kloudlite_params          = var.kloudlite_params
  tracker_id                = "${var.tracker_id}-masters"
  save_kubeconfig_to_path   = var.save_kubeconfig_to_path
  save_ssh_key_to_path      = var.save_ssh_key_to_path
  extra_server_args         = var.extra_server_args
}

module "kl-worker-nodes-on-aws" {
  source     = "../../terraform/bundles/kl-worker-nodes-on-aws"
  depends_on = [module.kl-master-nodes-on-aws.k3s_agent_token]
  aws_region = var.aws_region

  ec2_nodepools              = var.ec2_nodepools
  k3s_join_token             = module.kl-master-nodes-on-aws.k3s_agent_token
  k3s_server_public_dns_host = module.kl-master-nodes-on-aws.k3s_public_dns_host
  spot_nodepools             = var.spot_nodepools
  tracker_id                 = "${var.tracker_id}-workers"
  save_ssh_key_to_path       = var.save_worker_ssh_key_to_path
  extra_agent_args           = var.extra_agent_args
}
