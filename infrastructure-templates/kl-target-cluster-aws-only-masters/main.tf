module "kl-master-nodes-on-aws" {
  source                    = "../../terraform/bundles/kl-master-nodes-on-aws"
  aws_region                = var.aws_region
  enable_nvidia_gpu_support = var.enable_nvidia_gpu_support
  k3s_masters               = var.k3s_masters
  kloudlite_params          = var.kloudlite_params
  tracker_id                = "${var.tracker_id}-masters"
  save_ssh_key_to_path      = var.save_ssh_key_to_path
  save_kubeconfig_to_path   = var.save_kubeconfig_to_path
  extra_server_args         = var.extra_server_args
}
