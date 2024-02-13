locals {
  has_iam_instance_profile = var.aws_iam_instance_profile_role != ""
}

resource "aws_iam_instance_profile" "iam_instance_profile" {
  count = local.has_iam_instance_profile ? 1 : 0
  role  = var.aws_iam_instance_profile_role
}

module "aws-security-groups" {
  source                                = "../../../modules/aws/security-groups"
  allow_incoming_http_traffic_on_master = true
  allow_metrics_server_on_master        = true
  expose_k8s_node_ports_on_master       = true

  allow_incoming_http_traffic_on_agent    = false
  allow_metrics_server_on_agent           = true
  allow_outgoing_to_all_internet_on_agent = true
  expose_k8s_node_ports_on_agent          = false
}

locals {
  k3s_node_labels = {
    "kloudlite.io/cloud-provider.name" : "aws",
    "kloudlite.io/cloud-provider.region" : var.aws_region,
  }

  gpu_node_labels = {
    "kloudlite.io/node.has-gpu" : "true",
  }

  k3s_node_label_az = "kloudlite.io/cloud-provider.az"

  nodes_config = {
    for node_name, node_cfg in var.ec2_nodes_config : node_name => merge({
      iam_instance_profile = local.has_iam_instance_profile ? aws_iam_instance_profile.iam_instance_profile[0].name : null
      security_groups      = node_cfg.role == "agent" ? module.aws-security-groups.sg_for_k3s_agents_names : module.aws-security-groups.sg_for_k3s_masters_names
    }, node_cfg)
  }

  primary_master_nodes = {
    for node_name, node_cfg in local.nodes_config : node_name => node_cfg if node_cfg.role == "primary-master"
  }

  secondary_master_nodes = {
    for node_name, node_cfg in local.nodes_config : node_name => node_cfg if node_cfg.role == "secondary-master"
  }

  master_node_labels = merge({ "kloudlite.io/node.role" : "master" }, local.k3s_node_labels)
  agent_node_labels  = merge({ "kloudlite.io/node.role" : "agent" }, local.k3s_node_labels)

  agent_nodes = {for node_name, node_cfg in local.nodes_config : node_name => node_cfg if node_cfg.role == "agent"}

  spot_node_labels = merge({ "kloudlite.io/node.instance-type" : "spot" }, local.agent_node_labels)
}

locals {
  master_names = tolist(keys(merge(local.primary_master_nodes, local.secondary_master_nodes)))

  #  backup_crontab_schedule = {
  #    for idx, name in local.master_names : name => "*/1 * * * *"
  #  }

  backup_crontab_schedule = {
    for idx, name in local.master_names : name =>
    "* ${2 * (tonumber(idx) + 1)}/${2 * (length(local.primary_master_nodes) + length(local.secondary_master_nodes))} * * *"
  }
}

locals {
  master_node_taints = var.taint_master_nodes ?  {
    masters : {
      effect = "NoExecute"
      value  = "true"
    }
  } : {}
}

check "single_master" {
  assert {
    condition     = length(local.primary_master_nodes) == 1
    error_message = "only one primary master node is allowed"
  }
}

locals {
  primary_master_node_name = keys(local.primary_master_nodes)[0]
  primary_master_node      = local.primary_master_nodes[local.primary_master_node_name]
}

module "ec2-nodes" {
  source       = "../../../modules/aws/ec2-node"
  save_ssh_key = {
    enabled = true
    path    = "/tmp/ec2-ssh-key.pem"
  }

  ami            = var.aws_ami
  nvidia_gpu_ami = var.aws_nvidia_gpu_ami

  nodes_config = local.nodes_config
}

locals {
  masters_public_ip = [
    for node_name, v in merge(local.primary_master_nodes, local.secondary_master_nodes) :
    module.ec2-nodes.ec2_instances_public_ip[node_name]
  ]
}

module "k3s-primary-master" {
  source = "../../../modules/kloudlite/k3s/__deprecated__/k3s-primary-master"

  node_name           = local.primary_master_node_name
  public_dns_host = var.k3s_server_dns_hostname
  public_ip           = module.ec2-nodes.ec2_instanes_public_ip[local.primary_master_node_name]
  ssh_params          = {
    user        = var.aws_ami_ssh_username
    private_key = module.ec2-nodes.ssh_private_key
  }
  node_labels = merge({
    "kloudlite.io/cloud-provider.az" : local.primary_master_node.az,
  }, local.primary_master_node.is_nvidia_gpu_node == true ? local.gpu_node_labels : {}, local.master_node_labels)
  node_taints = local.master_node_taints

  k3s_master_nodes_public_ips = local.masters_public_ip
  backup_to_s3                = {
    enabled = var.k3s_backup_to_s3.enabled

    aws_access_key = var.aws_access_key
    aws_secret_key = var.aws_secret_key

    bucket_name   = var.k3s_backup_to_s3.bucket_name
    bucket_region = var.k3s_backup_to_s3.bucket_region
    bucket_folder = var.k3s_backup_to_s3.bucket_folder

    cron_schedule = local.backup_crontab_schedule[local.primary_master_node_name]
  }
  restore_from_latest_s3_snapshot = var.restore_from_latest_s3_snapshot
}

module "k3s-secondary-master" {
  source = "../../../modules/kloudlite/k3s/__deprecated__/k3s-secondary-master"

  k3s_token                = module.k3s-primary-master.k3s_token
  primary_master_public_ip = module.k3s-primary-master.public_ip
  public_dns_hostname      = var.k3s_server_dns_hostname

  depends_on = [module.k3s-primary-master]

  secondary_masters = {
    for node_name, node_cfg in local.secondary_master_nodes : node_name => {
      public_ip  = module.ec2-nodes.ec2_instances_public_ip[node_name]
      private_ip = module.ec2-nodes.ec2_instances_private_ip[node_name]
      ssh_params = {
        user        = var.aws_ami_ssh_username
        private_key = module.ec2-nodes.ssh_private_key
      }
      node_labels = merge({
        "kloudlite.io/cloud-provider.az" : node_cfg.az
      }, node_cfg.is_nvidia_gpu_node == true? local.gpu_node_labels : {}, local.master_node_labels)
      node_taints              = local.master_node_taints
      k3s_backup_cron_schedule = local.backup_crontab_schedule[node_name]
    }
  }
  k3s_master_nodes_public_ips = local.masters_public_ip

  backup_to_s3 = {
    enabled = var.k3s_backup_to_s3.enabled

    aws_access_key = var.aws_access_key
    aws_secret_key = var.aws_secret_key

    bucket_name   = var.k3s_backup_to_s3.bucket_name
    bucket_region = var.k3s_backup_to_s3.bucket_region
    bucket_folder = var.k3s_backup_to_s3.bucket_folder
  }

  restore_from_latest_s3_snapshot = var.restore_from_latest_s3_snapshot
}

module "cloudflare-dns" {
  count  = var.cloudflare.enabled ? 1 : 0
  source = "../../../modules/cloudflare/dns"

  cloudflare_api_token = var.cloudflare.api_token
  cloudflare_domain    = var.cloudflare.domain
  cloudflare_zone_id   = var.cloudflare.zone_id

  public_ips         = local.masters_public_ip
  set_wildcard_cname = true
}

module "k3s-agents" {
  source = "../../../modules/kloudlite/k3s/__deprecated__/k3s-agent"

  agent_nodes = {
    for node_name, node_cfg in local.agent_nodes : node_name => {
      public_ip  = module.ec2-nodes.ec2_instances_public_ip[node_name]
      ssh_params = {
        user        = var.aws_ami_ssh_username
        private_key = module.ec2-nodes.ssh_private_key
      }
      depends_on  = [module.k3s-primary-master]
      node_labels = merge({
        "kloudlite.io/cloud-provider.az" : node_cfg.az
      }, node_cfg.is_nvidia_gpu_node == true? local.gpu_node_labels : {}, local.agent_node_labels)
    }
  }

  use_cloudflare_nameserver = var.cloudflare.enabled
  k3s_server_dns_hostname   = var.k3s_server_dns_hostname
  k3s_token                 = module.k3s-primary-master.k3s_token
}

module "k3s-agents-on-aws-spot-fleets" {
  count  = var.spot_settings.enabled ? 1 : 0
  source = "../../../modules/kloudlite/k3s/__deprecated__/k3s-agents-on-aws-spot-fleets"

  aws_ami                 = var.aws_ami
  aws_nvidia_gpu_ami      = var.aws_nvidia_gpu_ami
  k3s_server_dns_hostname = var.k3s_server_dns_hostname
  k3s_token               = module.k3s-primary-master.k3s_token
  depends_on              = [module.k3s-primary-master]
  spot_nodes              = {
    for node_name, node_cfg in var.spot_nodes_config : node_name => merge(node_cfg, {
      security_groups      = module.aws-security-groups.sg_for_k3s_agents_ids
      iam_instance_profile = local.has_iam_instance_profile ? aws_iam_instance_profile.iam_instance_profile[0].name : null
      node_labels          = merge(
        local.spot_node_labels,
        node_cfg.az != "" ? { (local.k3s_node_label_az) : node_cfg.az } : {},
        node_cfg.nvidia_gpu.enabled ? local.gpu_node_labels : {},
      )
    })
  }
  save_ssh_key = {
    enabled = true
    path    = "/tmp/spot-ssh-key.pem"
  }
  disable_ssh                  = var.disable_ssh
  spot_fleet_tagging_role_name = var.spot_settings.spot_fleet_tagging_role_name
}

module "aws-k3s-spot-termination-handler" {
  count               = var.spot_settings.enabled ? 1 : 0
  source              = "../../../modules/kloudlite/spot-termination-handler"
  depends_on          = [module.k3s-primary-master]
  spot_nodes_selector = local.spot_node_labels
  ssh_params          = {
    public_ip   = module.k3s-primary-master.public_ip
    username    = var.aws_ami_ssh_username
    private_key = module.ec2-nodes.ssh_private_key
  }
}

module "kloudlite-crds" {
  count             = var.kloudlite.install_crds ? 1 : 0
  source            = "../../../modules/kloudlite/crds"
  kloudlite_release = var.kloudlite.release
  depends_on        = [module.k3s-primary-master]
  ssh_params        = {
    public_ip   = module.k3s-primary-master.public_ip
    username    = var.aws_ami_ssh_username
    private_key = module.ec2-nodes.ssh_private_key
  }
}

module "nvidia-container-runtime" {
  count      = var.enable_nvidia_gpu_support ? 1 : 0
  source     = "../../../modules/nvidia-container-runtime"
  depends_on = [module.kloudlite-crds]
  ssh_params = {
    public_ip   = module.ec2-nodes.ec2_instances_public_ip[local.primary_master_node_name]
    private_key = module.ec2-nodes.ssh_private_key
    user        = var.aws_ami_ssh_username
  }
  gpu_node_selector = local.gpu_node_labels
}

module "helm-aws-ebs-csi" {
  count           = var.kloudlite.install_csi_driver ? 1 : 0
  source          = "../../../modules/helm-charts/helm-aws-ebs-csi"
  depends_on      = [module.kloudlite-crds]
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
  node_selector = local.agent_node_labels
  ssh_params    = {
    public_ip   = module.k3s-primary-master.public_ip
    username    = var.aws_ami_ssh_username
    private_key = module.ec2-nodes.ssh_private_key
  }
}

module "kloudlite-operators" {
  count             = var.kloudlite.install_operators ? 1 : 0
  source            = "../../../modules/kloudlite/helm-kloudlite-operators"
  depends_on        = [module.kloudlite-crds]
  kloudlite_release = var.kloudlite.release
  node_selector     = {}
  ssh_params        = {
    public_ip   = module.k3s-primary-master.public_ip
    username    = var.aws_ami_ssh_username
    private_key = module.ec2-nodes.ssh_private_key
  }
}

module "kloudlite-agent" {
  count                              = var.kloudlite.install_agent ? 1 : 0
  source                             = "../../../modules/kloudlite/helm-kloudlite-agent"
  kloudlite_account_name             = var.kloudlite.agent_vars.account_name
  kloudlite_cluster_name             = var.kloudlite.agent_vars.cluster_name
  kloudlite_cluster_token            = var.kloudlite.agent_vars.cluster_token
  kloudlite_dns_host                 = var.kloudlite.agent_vars.dns_host
  kloudlite_message_office_grpc_addr = var.kloudlite.agent_vars.message_office_grpc_addr
  kloudlite_release                  = var.kloudlite.release
  ssh_params                         = {
    public_ip   = module.k3s-primary-master.public_ip
    username    = var.aws_ami_ssh_username
    private_key = module.ec2-nodes.ssh_private_key
  }
}

module "disable_ssh_on_instances" {
  count      = var.disable_ssh ? 1 : 0
  source     = "../../../modules/disable-ssh-on-nodes"
  depends_on = [
    module.k3s-primary-master,
    module.k3s-secondary-master,
    module.k3s-agents,
  ]
  nodes_config = {
    for name, config in local.nodes_config : name => {
      public_ip  = module.ec2-nodes.ec2_instances_public_ip[name]
      ssh_params = {
        user        = var.aws_ami_ssh_username
        private_key = module.ec2-nodes.ssh_private_key
      }
      disable_ssh = var.disable_ssh
    }
  }
}
