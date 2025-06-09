module "ssh-rsa-key" {
  source = "../../../modules/common/ssh-rsa-key"
}

resource "aws_key_pair" "ssh_key_pair" {
  key_name   = "${var.trace_id}-ssh-key-pair"
  public_key = module.ssh-rsa-key.public_key
}

module "ec2-node" {
  source   = "../../../modules/aws/ec2-node-v2"
  for_each = [for node in var.nodes : node]
  trace_id = var.trace_id

  name = each.value.name

  ami    = each.value.ami
  vpc_id = var.vpc_id

  instance_type  = each.value.instance_type
  instance_state = var.instance_state

  availability_zone    = each.value.availability_zone
  iam_instance_profile = var.iam_instance_profile
  root_volume_size     = each.value.root_volume_size
  root_volume_type     = each.value.root_volume_type
  ssh_key_name         = aws_key_pair.ssh_key_pair.key_name

  user_data_base64 = base64encode(
    templatefile("${path.module}/launch-script.sh", {
      k3s_server_host = var.k3s_server_host
      k3s_agent_token = var.k3s_agent_token
      k3s_version     = var.k3s_version
      node_name       = var.name,
  }))

  security_group_ids = each.value.security_group_ids
  subnet_id          = each.value.vpc_subnet_id
}

module "aws-security-groups" {
  source     = "../../../modules/aws/security-groups"
  depends_on = [null_resource.variable_validations]

  tracker_id = var.tracker_id
  vpc_id     = var.vpc_id

  create_for_k3s_masters = true

  allow_incoming_http_traffic = true
  expose_k8s_node_ports       = true
}

module "kloudlite-k3s-templates" {
  source = "../../../modules/kloudlite/k3s/k3s-templates"
}

module "aws-amis" {
  source = "../../../modules/aws/AMIs"
}

module "k3s-master-instances" {
  source   = "../../../modules/aws/ec2-node"
  for_each = { for name, cfg in var.k3s_masters.nodes : name => cfg }

  ami           = var.k3s_masters.ami
  instance_type = var.k3s_masters.instance_type

  availability_zone    = each.value.availability_zone
  iam_instance_profile = var.k3s_masters.iam_instance_profile
  is_nvidia_gpu_node   = var.enable_nvidia_gpu_support
  node_name            = each.key
  root_volume_size     = var.k3s_masters.root_volume_size
  root_volume_type     = var.k3s_masters.root_volume_type
  security_groups      = module.aws-security-groups.sg_for_k3s_masters_names
  last_recreated_at    = each.value.last_recreated_at
  ssh_key_name         = aws_key_pair.k3s_nodes_ssh_key.key_name
  tracker_id           = var.tracker_id
  tags                 = var.tags
  user_data_base64 = base64encode(templatefile(module.kloudlite-k3s-templates.k3s-vm-setup-template-path, {
    kloudlite_release             = var.kloudlite_params.release
    k3s_download_url              = ""
    kloudlite_runner_download_url = ""
    kloudlite_config_directory    = module.kloudlite-k3s-templates.kloudlite_config_directory
  }))
  vpc = {
    subnet_id              = each.value.vpc_subnet_id
    vpc_security_group_ids = module.aws-security-groups.sg_for_k3s_masters_ids
  }
}

module "kloudlite-k3s-masters" {
  source                    = "../../kloudlite-k3s-masters"
  backup_to_s3              = var.k3s_masters.backup_to_s3
  cloudflare                = var.k3s_masters.cloudflare
  cluster_internal_dns_host = var.k3s_masters.cluster_internal_dns_host
  enable_nvidia_gpu_support = var.enable_nvidia_gpu_support
  kloudlite_params          = var.kloudlite_params
  master_nodes = {
    for name, cfg in var.k3s_masters.nodes : name => {
      role : cfg.role,
      public_ip : module.k3s-master-instances[name].public_ip,
      node_labels : {
        (module.constants.node_labels.kloudlite_release) : cfg.kloudlite_release,
        (module.constants.node_labels.provider_name) : "aws",
        (module.constants.node_labels.provider_region) : var.aws_region,
        (module.constants.node_labels.provider_az) : cfg.availability_zone,
        (module.constants.node_labels.node_has_role) : cfg.role,
        (module.constants.node_labels.provider_aws_instance_profile_name) : var.k3s_masters.iam_instance_profile,
        (module.constants.node_labels.provider_instance_type) : var.k3s_masters.instance_type,
      }
      availability_zone = cfg.availability_zone,
      last_recreated_at = cfg.last_recreated_at,
      kloudlite_release = cfg.kloudlite_release,
    }
  }
  public_dns_host              = var.k3s_masters.public_dns_host
  restore_from_latest_snapshot = var.k3s_masters.restore_from_latest_snapshot
  ssh_private_key              = module.ssh-rsa-key.private_key

  ssh_username = var.k3s_masters.ssh_username

  taint_master_nodes      = var.k3s_masters.taint_master_nodes
  tracker_id              = var.tracker_id
  save_kubeconfig_to_path = var.save_kubeconfig_to_path
  cloudprovider_name      = "aws"
  cloudprovider_region    = var.aws_region
  k3s_service_cidr        = var.k3s_service_cidr
}
