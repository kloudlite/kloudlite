locals {
  primary_master_node_name = one([
    for node_name, node_cfg in var.k3s_masters.nodes : node_name
    if node_cfg.role == "primary-master"
  ])
}

resource "null_resource" "variable_validations" {
  lifecycle {
    precondition {
      error_message = "k3s_masters.nodes can/must have only one node with role primary-master"
      condition     = local.primary_master_node_name != null
    }
  }
}

module "constants" {
  source = "../../modules/common/constants"
}

module "ssh-rsa-key" {
  source     = "../../modules/common/ssh-rsa-key"
  depends_on = [null_resource.variable_validations]
}

resource "null_resource" "save_ssh_key" {
  count = length(var.save_ssh_key_to_path) > 0? 1 : 0

  provisioner "local-exec" {
    quiet   = true
    command = "echo '${module.ssh-rsa-key.private_key}' > ${var.save_ssh_key_to_path} && chmod 600 ${var.save_ssh_key_to_path}"
  }
}

resource "random_id" "id" {
  byte_length = 4
  depends_on  = [null_resource.variable_validations]
}

resource "aws_key_pair" "k3s_nodes_ssh_key" {
  key_name   = "${var.tracker_id}-${random_id.id.hex}-ssh-key"
  public_key = module.ssh-rsa-key.public_key
  depends_on = [null_resource.variable_validations]
}

module "aws-security-groups" {
  source                                = "../../modules/aws/security-groups"
  depends_on                            = [null_resource.variable_validations]
  create_group_for_k3s_masters          = true
  allow_incoming_http_traffic_on_master = true
  allow_metrics_server_on_master        = true
  expose_k8s_node_ports_on_master       = true

  allow_incoming_http_traffic_on_agent    = false
  allow_metrics_server_on_agent           = true
  allow_outgoing_to_all_internet_on_agent = true
  expose_k8s_node_ports_on_agent          = false
  tracker_id                              = var.tracker_id
}

module "k3s-master-instances" {
  source             = "../../modules/aws/aws-ec2-nodepool"
  depends_on         = [null_resource.variable_validations]
  instance_type      = var.k3s_masters.instance_type
  nodes              = {for name, cfg in var.k3s_masters.nodes : name => { last_recreated_at : cfg.last_recreated_at }}
  nvidia_gpu_enabled = var.k3s_masters.nvidia_gpu_enabled
  root_volume_size   = var.k3s_masters.root_volume_size
  root_volume_type   = var.k3s_masters.root_volume_type
  security_groups    = module.aws-security-groups.sg_for_k3s_masters_names
  ssh_key_name       = aws_key_pair.k3s_nodes_ssh_key.key_name
  tracker_id         = "${var.tracker_id}-masters"
  ami                = var.k3s_masters.image_id
}

module "kloudlite-k3s-masters" {
  source                    = "../kloudlite-k3s-masters"
  backup_to_s3              = var.k3s_masters.backup_to_s3
  cloudflare                = var.k3s_masters.cloudflare
  cluster_internal_dns_host = var.k3s_masters.cluster_internal_dns_host
  enable_nvidia_gpu_support = var.enable_nvidia_gpu_support
  kloudlite_params          = var.kloudlite_params
  master_nodes              = {
    for name, cfg in var.k3s_masters.nodes : name => {
      role : cfg.role,
      public_ip : module.k3s-master-instances.public_ips[name],
      node_labels : {
        (module.constants.node_labels.provider_name) : "aws",
        (module.constants.node_labels.provider_region) : var.aws_region,
        (module.constants.node_labels.provider_az) : cfg.availability_zone,
        (module.constants.node_labels.node_has_role) : cfg.role,
        (module.constants.node_labels.provider_aws_instance_profile_name) : var.k3s_masters.iam_instance_profile,
      }
      availability_zone = cfg.availability_zone,
      last_recreated_at = cfg.last_recreated_at,
    }
  }
  public_dns_host              = var.k3s_masters.public_dns_host
  restore_from_latest_snapshot = var.k3s_masters.restore_from_latest_snapshot
  ssh_private_key              = module.ssh-rsa-key.private_key
  ssh_username                 = var.k3s_masters.image_ssh_username
  taint_master_nodes           = var.k3s_masters.taint_master_nodes
  tracker_id                   = var.tracker_id
  save_kubeconfig_to_path      = var.save_kubeconfig_to_path
}

module "helm-aws-ebs-csi" {
  count           = var.kloudlite_params.install_csi_driver ? 1 : 0
  source          = "../../modules/helm-charts/helm-aws-ebs-csi"
  depends_on      = [module.kloudlite-k3s-masters]
  storage_classes = {
    "sc-xfs" : {
      volume_type = "gp3"
      fs_type     = "xfs"
    },
    "sc-ext4" : {
      volume_type = "gp3"
      fs_type     = "ext4"
    },
  }
  controller_node_selector = module.constants.master_node_selector
  controller_tolerations   = module.constants.master_node_tolerations
  daemonset_node_selector  = module.constants.agent_node_selector
  ssh_params               = {
    public_ip   = module.kloudlite-k3s-masters.k3s_primary_master_public_ip
    username    = var.k3s_masters.image_ssh_username
    private_key = module.ssh-rsa-key.private_key
  }
}

module "aws-k3s-spot-termination-handler" {
  source              = "../../modules/kloudlite/spot-termination-handler"
  depends_on          = [module.kloudlite-k3s-masters.kubeconfig]
  spot_nodes_selector = module.constants.spot_node_selector
  ssh_params          = {
    public_ip   = module.kloudlite-k3s-masters.k3s_primary_master_public_ip
    username    = var.k3s_masters.image_ssh_username
    private_key = module.ssh-rsa-key.private_key
  }
  kloudlite_release = var.kloudlite_params.release
  release_name      = "kl-aws-spot-termination-handler"
  release_namespace = module.kloudlite-k3s-masters.kloudlite_namespace
}