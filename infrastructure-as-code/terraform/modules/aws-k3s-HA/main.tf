locals {
  has_iam_instance_profile = var.aws_iam_instance_profile_role != ""
}

resource "aws_iam_instance_profile" "iam_instance_profile" {
  count = local.has_iam_instance_profile ? 1 : 0
  role  = var.aws_iam_instance_profile_role
}

module "aws-security-groups" {
  source = "../aws-security-groups"
}

locals {
  k3s_node_labels = {
    "kloudlite.io/cloud-provider.name" : "aws",
    "kloudlite.io/cloud-provider.region" : var.aws_region,
  }

  k3s_node_label_az = "kloudlite.io/cloud-provider.az"

  nodes_config = {
    for node_name, node_cfg in var.ec2_nodes_config : node_name => merge({
      iam_instance_profile = local.has_iam_instance_profile ? aws_iam_instance_profile.iam_instance_profile[0].name : null
      security_groups      = node_cfg.role == "agent" ? module.aws-security-groups.security_group_k3s_agents_names : module.aws-security-groups.security_group_k3s_masters_names
    }, node_cfg)
  }

  primary_master_nodes = {
    for node_name, node_cfg in local.nodes_config : node_name => node_cfg if node_cfg.role == "primary-master"
  }

  secondary_master_nodes = {
    for node_name, node_cfg in local.nodes_config : node_name => node_cfg if node_cfg.role == "secondary-master"
  }

  agent_nodes = {for node_name, node_cfg in local.nodes_config : node_name => node_cfg if node_cfg.role == "agent"}

  spot_node_labels = merge({ "kloudlite.io/node-instance-type" : "spot" }, local.k3s_node_labels)
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
  source       = "../ec2-nodes"
  save_ssh_key = {
    enabled = true
    path    = "/tmp/ec2-ssh-key.pem"
  }

  aws_access_key = var.aws_access_key
  aws_secret_key = var.aws_secret_key

  ami        = var.aws_ami
  aws_region = var.aws_region

  nodes_config = local.nodes_config
}

locals {
  public_ips = [
    for node_name, v in merge(local.primary_master_nodes, local.secondary_master_nodes) :
    module.ec2-nodes.ec2_instances_public_ip[node_name]
  ]
}

module "k3s-primary-master" {
  source = "../k3s-primary-master"

  node_name           = local.primary_master_node_name
  public_dns_hostname = var.k3s_server_dns_hostname
  public_ip           = module.ec2-nodes.ec2_instances_public_ip[local.primary_master_node_name]
  ssh_params          = {
    user        = var.aws_ami_ssh_username
    private_key = module.ec2-nodes.ssh_private_key
  }
  node_labels = merge({
    "kloudlite.io/cloud-provider.az" : local.primary_master_nodes[local.primary_master_node_name].az
  }, local.k3s_node_labels)

  k3s_master_nodes_public_ips = local.public_ips
}

module "k3s-secondary-master" {
  source = "../k3s-secondary-master"

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
      node_labels = merge({ "kloudlite.io/cloud-provider.az" : node_cfg.az }, local.k3s_node_labels)
    }
  }
  k3s_master_nodes_public_ips = local.public_ips
}

module "k3s-agents" {
  source = "../k3s-agents"

  agent_nodes = {
    for node_name, node_cfg in local.agent_nodes : node_name => {
      public_ip  = module.ec2-nodes.ec2_instances_public_ip[node_name]
      ssh_params = {
        user        = var.aws_ami_ssh_username
        private_key = module.ec2-nodes.ssh_private_key
      }
      depends_on  = [module.k3s-primary-master]
      node_labels = merge({ "kloudlite.io/cloud-provider.az" : node_cfg.az }, local.k3s_node_labels)
    }
  }

  use_cloudflare_nameserver = var.cloudflare.enabled
  k3s_server_dns_hostname   = var.k3s_server_dns_hostname
  k3s_token                 = module.k3s-primary-master.k3s_token
}

module "cloudflare-dns" {
  count  = var.cloudflare.enabled ? 1 : 0
  source = "../cloudflare-dns"

  cloudflare_api_token = var.cloudflare.api_token
  cloudflare_domain    = var.cloudflare.domain
  cloudflare_zone_id   = var.cloudflare.zone_id

  public_ips = local.public_ips
}

module "helm-aws-ebs-csi" {
  source          = "../helm-charts/aws-ebs-csi"
  kubeconfig      = module.k3s-primary-master.kubeconfig_with_public_ip
  depends_on      = [module.k3s-primary-master]
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
  node_selector = {}
}

module "k3s-agents-on-ec2-fleets" {
  source = "../k3s-agents-on-ec2-fleets"

  aws_ami                 = var.aws_ami
  k3s_server_dns_hostname = var.k3s_server_dns_hostname
  k3s_token               = module.k3s-primary-master.k3s_token
  depends_on              = [module.k3s-primary-master]
  #  spot_nodes      = {}
  spot_nodes              = {
    for node_name, node_cfg in var.spot_nodes_config : node_name => {
      instance_type        = node_cfg.instance_type
      az                   = node_cfg.az
      security_groups      = module.aws-security-groups.security_group_k3s_agents_ids
      iam_instance_profile = local.has_iam_instance_profile ? aws_iam_instance_profile.iam_instance_profile[0].name : null
      node_labels          = merge({
        "kloudlite.io/cloud-provider.az" : node_cfg.az,
      }, local.spot_node_labels)
    }
  }
  save_ssh_key = {
    enabled = true
    path    = "/tmp/spot-ssh-key.pem"
  }
  spot_fleet_tagging_role_name = var.spot_settings.spot_fleet_tagging_role_name
}

module "disable_ssh_on_instances" {
  source     = "../disable-ssh-on-nodes"
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

module "aws-k3s-spot-termination-handler" {
  source              = "../aws-k3s-spot-termination-handler"
  depends_on          = [module.k3s-primary-master]
  spot_nodes_selector = local.spot_node_labels
}