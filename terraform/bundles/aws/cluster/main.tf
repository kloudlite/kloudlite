module "ssh-rsa-key" {
  source = "../../../modules/common/ssh-rsa-key"
}

resource "aws_key_pair" "ssh_key_pair" {
  key_name   = "${var.cluster_name}-ssh-key-pair"
  public_key = module.ssh-rsa-key.public_key
}

# resource "aws_iam_role" "instance_role" {
#   name = "InstanceRole"
#
#   assume_role_policy = jsonencode({
#     Version = "2012-10-17"
#     Statement = [{
#       Effect = "Allow"
#       Principal = {
#         Service = "ec2.amazonaws.com"
#       }
#       Action = "sts:AssumeRole"
#     }]
#   })
# }

# resource "aws_iam_role_policy_attachments_exclusive" "instance_role_managed_policy_arns" {
#   role_name   = aws_iam_role.instance_role.name
#   policy_arns = ["arn:aws:iam::aws:policy/AmazonEC2FullAccess", "arn:aws:iam::aws:policy/AmazonS3FullAccess"]
# }
#
# resource "aws_iam_instance_profile" "instance_profile" {
#   name = "InstanceProfile"
#   role = aws_iam_role.instance_role.name
# }

resource "random_password" "k3s_server_token" {
  length           = 24
  special          = true
  upper            = true
  lower            = true
  numeric          = true
  override_special = "!@#%^&*"
}

resource "random_password" "k3s_agent_token" {
  length           = 24
  special          = true
  upper            = true
  lower            = true
  numeric          = true
  override_special = "!@#%^&*"
}

data "aws_vpc" "current" {
  id = var.vpc_id
}


resource "aws_security_group" "nlb_sg" {
  name        = "NLBSecurityGroup-${var.cluster_name}"
  description = "Public access to NLB for HTTP and HTTPS"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "cluster_sg" {
  name        = "ClusterSecurityGroup-${var.cluster_name}"
  description = "Internal security group for kloudlite cluster"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 443
    to_port         = 443
    protocol        = "tcp"
    security_groups = [aws_security_group.nlb_sg.id]
  }

  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.current.cidr_block]
  }

  # etcd
  ingress {
    from_port   = 2379
    to_port     = 2379
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.current.cidr_block]
  }

  ingress {
    from_port   = 2380
    to_port     = 2380
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.current.cidr_block]
  }

  # flannel wireguard-native
  ingress {
    from_port   = 51820
    to_port     = 51820
    protocol    = "udp"
    cidr_blocks = [data.aws_vpc.current.cidr_block]
  }

  # Kubelet metrics
  ingress {
    from_port   = 10250
    to_port     = 10250
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.current.cidr_block]
  }

  # SSH (internal to VPC, for debugging purposes)
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.current.cidr_block]
  }
}

module "ec2-nodes" {
  source   = "../../../modules/aws/ec2-node-v2"
  for_each = { for node in var.master_nodes : node.name => node }
  trace_id = var.cluster_name

  name   = each.value.name
  ami    = each.value.ami
  vpc_id = var.vpc_id

  instance_type  = each.value.instance_type
  instance_state = var.cluster_state

  availability_zone = each.value.availability_zone
  # iam_instance_profile = aws_iam_instance_profile.instance_profile.name
  # iam_instance_profile = var.master_node_iam_instance_profile
  root_volume_size = each.value.root_volume_size
  root_volume_type = each.value.root_volume_type
  ssh_key_name     = aws_key_pair.ssh_key_pair.key_name

  user_data_base64 = base64encode(
    templatefile("${path.module}/launch-script.sh", {
      k3s_server_token = random_password.k3s_server_token.result
      k3s_agent_token  = random_password.k3s_agent_token.result
      k3s_version      = each.value.k3s_version
      node_name        = each.value.name

      kloudlite_release = var.kloudlite_release
      base_domain       = var.base_domain
  }))

  security_group_ids = [aws_security_group.cluster_sg.id]
  subnet_id          = each.value.vpc_subnet_id
}

resource "aws_lb" "network_load_balancer" {
  name               = "${var.cluster_name}-nlb"
  internal           = false
  load_balancer_type = "network"
  subnets            = toset([for node in var.master_nodes : node.vpc_subnet_id])
  security_groups    = [aws_security_group.nlb_sg.id]

  tags = {
    Name = "${var.cluster_name}-nlb"
  }
}

resource "aws_lb_target_group" "http" {
  for_each    = { for node in var.master_nodes : node.name => node }
  name        = "${var.cluster_name}-${each.key}-tg-http"
  port        = 80
  protocol    = "TCP"
  vpc_id      = var.vpc_id
  target_type = "instance"

  health_check {
    protocol = "TCP"
  }
}

resource "aws_lb_target_group" "https" {
  for_each    = { for node in var.master_nodes : node.name => node }
  name        = "${var.cluster_name}-${each.key}-tg-https"
  port        = 443
  protocol    = "TCP"
  vpc_id      = var.vpc_id
  target_type = "instance"

  health_check {
    protocol = "TCP"
  }
}

resource "aws_lb_target_group_attachment" "http-attachment" {
  for_each         = module.ec2-nodes
  target_group_arn = aws_lb_target_group.http[each.key].arn
  target_id        = each.value.instance_id
  port             = 80
}

resource "aws_lb_target_group_attachment" "https-attachment" {
  for_each         = module.ec2-nodes
  target_group_arn = aws_lb_target_group.https[each.key].arn
  target_id        = each.value.instance_id
  port             = 443
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.network_load_balancer.arn
  port              = 80
  protocol          = "TCP"

  dynamic "default_action" {
    for_each = { for k, v in aws_lb_target_group.http : k => v.arn }
    content {
      type             = "forward"
      target_group_arn = default_action.value
    }
  }
}

resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.network_load_balancer.arn
  port              = 443
  protocol          = "TCP"

  dynamic "default_action" {
    for_each = { for k, v in aws_lb_target_group.http : k => v.arn }
    content {
      type             = "forward"
      target_group_arn = default_action.value
    }
  }
}

