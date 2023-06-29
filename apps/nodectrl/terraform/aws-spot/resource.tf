terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.3.0"
    }
  }
  required_version = ">= 1.2.0"
}

provider "aws" {
  region     = var.region
  access_key = var.access_key
  secret_key = var.secret_key
}

output "node-name" {
  value = var.node_name
}


data "aws_caller_identity" "current" {}

resource "aws_security_group" "sg" {

  name = "sg-${var.node_name}"

  ingress {
    from_port   = 22
    protocol    = "tcp"
    to_port     = 22
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 2379
    protocol    = "tcp"
    to_port     = 2379
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 2380
    protocol    = "tcp"
    to_port     = 2380
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 6443
    protocol    = "tcp"
    to_port     = 6443
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 8472
    protocol    = "udp"
    to_port     = 8472
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 9100
    protocol    = "tcp"
    to_port     = 9100
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 51820
    protocol    = "udp"
    to_port     = 51820
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 51821
    protocol    = "udp"
    to_port     = 51821
    cidr_blocks = ["0.0.0.0/0"]
  }


  ingress {
    from_port   = 10250
    protocol    = "tcp"
    to_port     = 10250
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    protocol    = "tcp"
    to_port     = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    protocol    = "tcp"
    to_port     = 443
    cidr_blocks = ["0.0.0.0/0"]
  }


  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

}


resource "aws_launch_template" "spot-template" {
  name     = var.node_name
  image_id = "ami-0e63f370aa626048d"


  user_data = base64encode(templatefile("./init.sh", {
    pubkey         = file("${var.keys_path}/access.pub")
    nodeConfigYaml = file("${var.keys_path}/data.yaml")
  }))



  block_device_mappings {
    device_name = "/dev/sda1"
    ebs {
      volume_size = 40
    }
  }

  network_interfaces {
    associate_public_ip_address = true
    security_groups             = [aws_security_group.sg.id]
  }

  tag_specifications {
    resource_type = "instance"
    tags = {
      Name = var.node_name
    }
  }
}



resource "aws_spot_fleet_request" "byoc-spot-node" {
  iam_fleet_role = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-ec2-spot-fleet-tagging-role"

  target_capacity = 1

  terminate_instances_on_delete = true
  on_demand_target_capacity     = 0
  allocation_strategy           = "priceCapacityOptimized"
  on_demand_allocation_strategy = "lowestPrice"


  launch_template_config {
    launch_template_specification {
      id      = aws_launch_template.spot-template.id
      version = "1"
    }
    overrides {
      instance_requirements {
        vcpu_count {
          min = var.cpu_min
          max = var.cpu_max
        }
        memory_mib {
          min = var.mem_min
          max = var.mem_max
        }
      }
    }
  }
}
