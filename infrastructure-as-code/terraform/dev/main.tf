resource "aws_iam_instance_profile" "full_s3_and_block_storage" {
  role = "EC2StorageAccess"
}

module "aws-security-groups" {
  source = "../modules/aws/security-groups"
}


locals {
  k3s_node_labels = {
    "kloudlite.io/cloud-provider.name" : "aws",
    "kloudlite.io/cloud-provider.region" : var.aws_region,
  }

  k3s_node_label_az = "kloudlite.io/cloud-provider.az"

  nodes_config = {
    for node_name, node_cfg in var.nodes_config : node_name => merge({
      iam_instance_profile = aws_iam_instance_profile.full_s3_and_block_storage.name
      security_groups      = node_cfg.role == "agent" ? module.aws-security-groups.sg_for_k3s_agents_names : module.aws-security-groups.sg_for_k3s_masters_names
    }, node_cfg)
  }

  primary_master_nodes = {
    for node_name, node_cfg in local.nodes_config : node_name => node_cfg if node_cfg.role == "primary-master"
  }
  secondary_master_nodes = {
    for node_name, node_cfg in local.nodes_config : node_name => node_cfg if node_cfg.role == "secondary-master"
  }
  agent_nodes = {for node_name, node_cfg in local.nodes_config : node_name => node_cfg if node_cfg.role == "agent"}

  spot_node_labels = merge({
    "kloudlite.io/node-instance-type" : "spot"
  }, local.k3s_node_labels)
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
  source       = "../modules/ec2-nodes"
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
    module.ec2-nodes.ec2_instances[node_name].public_ip
  ]
}

module "k3s-primary-master" {
  source = "../modules/k3s/k3s-primary-master"

  node_name     = local.primary_master_node_name
  public_dns_hostname = var.cloudflare_domain
  public_ip     = module.ec2-nodes.ec2_instances[local.primary_master_node_name].public_ip
  ssh_params    = {
    user        = "ubuntu",
    private_key = module.ec2-nodes.ssh_private_key
  }
  node_labels = merge({
    "kloudlite.io/cloud-provider.az" : local.primary_master_node.az
  }, local.k3s_node_labels)
  use_cloudflare_nameserver = true

  disable_ssh                 = false
  k3s_master_nodes_public_ips = local.public_ips
}

module "k3s-secondary-master" {
  source = "../modules/k3s/k3s-secondary-master"

  k3s_token                = module.k3s-primary-master.k3s_token
  primary_master_public_ip = module.k3s-primary-master.public_ip
  public_dns_hostname            = var.cloudflare_domain

  depends_on = [module.k3s-primary-master]

  secondary_masters = {
    for node_name, node_cfg in local.secondary_master_nodes : node_name => {
      public_ip  = module.ec2-nodes.ec2_instances[node_name].public_ip
      private_ip = module.ec2-nodes.ec2_instances[node_name].private_ip
      ssh_params = {
        user        = "ubuntu"
        private_key = module.ec2-nodes.ssh_private_key
      }
      node_labels = merge({ "kloudlite.io/cloud-provider.az" : node_cfg.az }, local.k3s_node_labels)
    }
  }
  k3s_master_nodes_public_ips = local.public_ips
}

module "k3s-agents" {
  source = "../modules/k3s/k3s-agents"

  agent_nodes = {
    for node_name, node_cfg in local.agent_nodes : node_name => {
      public_ip  = module.ec2-nodes.ec2_instances_public_ip[node_name]
      ssh_params = {
        user        = "ubuntu"
        private_key = module.ec2-nodes.ssh_private_key
      }
      depends_on  = [module.k3s-primary-master]
      node_labels = merge({ "kloudlite.io/cloud-provider.az" : node_cfg.az }, local.k3s_node_labels)
    }
  }

  use_cloudflare_nameserver = true

  k3s_server_dns_hostname = var.cloudflare_domain
  k3s_token       = module.k3s-primary-master.k3s_token
}

output "ec2_instances" {
  value = module.ec2-nodes.ec2_instances
}

output "kubeconfig" {
  value = module.k3s-primary-master.kubeconfig_with_public_host
}


module "cloudflare-dns" {
  source = "../modules/cloudflare/dns"

  cloudflare_api_token = var.cloudflare_api_token
  cloudflare_domain    = var.cloudflare_domain
  cloudflare_zone_id   = var.cloudflare_zone_id

  public_ips = local.public_ips
}

module "helm-aws-ebs-csi" {
  source          = "../modules/helm-charts/helm-aws-ebs-csi"
  kubeconfig      = module.k3s-primary-master.kubeconfig_with_public_ip
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
}

module "k3s-agents-on-ec2-fleets" {
  source = "../modules/k3s/k3s-agents-on-ec2-fleets"

  providers = {
    aws = aws
  }

  aws_ami         = var.aws_ami
  k3s_server_dns_hostname = var.cloudflare_domain
  k3s_token       = module.k3s-primary-master.k3s_token
  depends_on      = [module.k3s-primary-master]
  spot_nodes      = {
    for node_name, node_cfg in var.spot_nodes_config : node_name => {
      instance_type        = node_cfg.instance_type
      az                   = node_cfg.az
      security_groups      = module.aws-security-groups.sg_for_k3s_agents_ids
      iam_instance_profile = aws_iam_instance_profile.full_s3_and_block_storage.name
      node_labels          = merge({
        "kloudlite.io/cloud-provider.az" : node_cfg.az,
      }, local.spot_node_labels)
    }
  }
  save_ssh_key = {
    enabled = true
    path    = "/tmp/spot-ssh-key.pem"
  }
  spot_fleet_tagging_role_name = "aws-ec2-spot-fleet-tagging-role"
}

module "disable_ssh_on_instances" {
  source     = "../modules/disable-ssh-on-nodes"
  depends_on = [
    module.k3s-primary-master,
    module.k3s-secondary-master,
    module.k3s-agents,
  ]
  nodes_config = {
    for name, config in local.nodes_config : name => {
      public_ip  = module.ec2-nodes.ec2_instances[name].public_ip
      ssh_params = {
        user        = "ubuntu"
        private_key = module.ec2-nodes.ssh_private_key
      }
      disable_ssh = var.disable_ssh
    }
  }
}

module "aws-k3s-spot-termination-handler" {
  source = "../modules/aws/spot-termination-handler"
  spot_nodes_selector = local.spot_node_labels
}