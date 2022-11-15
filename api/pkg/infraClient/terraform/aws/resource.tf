terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.16"
    }
  }

  required_version = ">= 1.2.0"
}

provider "aws" {
  region  = var.region
  access_key = var.access_key
  secret_key = var.secret_key
}


data "aws_ami" "latest-ubuntu" {
  most_recent = true
  owners = ["099720109477"] # Canonical

  filter {
      name   = "name"
      values = ["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"]
  }

  filter {
      name   = "virtualization-type"
      values = ["hvm"]
  }
}


resource "aws_security_group" "sg" {
  # name = "${var.node_id}-sg"

  ingress {
    from_port = 22
    protocol = "tcp"
    to_port = 22
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  ingress {
    from_port = 6443
    protocol = "tcp"
    to_port = 6443
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 8472
    protocol = "udp"
    to_port = 8472
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 9100
    protocol = "tcp"
    to_port = 9100
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 51820
    protocol = "udp"
    to_port = 51820
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 51821
    protocol = "udp"
    to_port = 51821
    cidr_blocks = ["0.0.0.0/0"]
  }


  ingress {
    from_port = 10250
    protocol = "tcp"
    to_port = 10250
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 80
    protocol = "tcp"
    to_port = 80
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 443
    protocol = "tcp"
    to_port = 443
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 30000
    protocol = "tcp"
    to_port = 32768
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 30000
    protocol = "udp"
    to_port = 32768
    cidr_blocks = ["0.0.0.0/0"]
  }


  egress {
    from_port       = 0
    to_port         = 0
    protocol        = "-1"
    cidr_blocks     = ["0.0.0.0/0"]
  }

  # lifecycle {
  #   create_before_destroy = true
  # }
}



resource "aws_instance" "byoc-node" {
  ami           = "${data.aws_ami.latest-ubuntu.id}"
  instance_type = var.instance_type
  security_groups = [aws_security_group.sg.name]

  user_data = templatefile("./init.sh", {
    pubkey = file("${var.keys-path}/access.pub")
    # hostname = var.node_id
  })

  tags = {
    Name = var.node_id
  }
}


output "node-ip" {
  value =  aws_instance.byoc-node.public_ip
}

output "node-name" {
  value =  var.node_id
}
