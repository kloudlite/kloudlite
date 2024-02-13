resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"
  tags       = merge(var.tags, {
    Name = "${var.tracker_id}-${var.vpc_name}"
  })
}

data "aws_availability_zones" "az" {
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

#locals {
#  public_subnet = {for idx, subnet in var.public_subnets : subnet.availability_zone => subnet.cidr}
#}

resource "aws_subnet" "public_subnets" {
  for_each = {
    for idx, value in data.aws_availability_zones.az.names : idx => {
      name = value
      cidr = "10.0.${idx+1}.0/24"
    }
  }
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = each.value.cidr
  availability_zone       = each.value.name
  map_public_ip_on_launch = true
  tags                    = merge(var.tags, {
    Name = "${var.tracker_id}-${each.value.name}-public"
  })
}

resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = aws_vpc.vpc.id
  tags   = merge(var.tags, {
    Name = "${var.tracker_id}-ingress-gateway"
  })
}

resource "aws_route_table" "public_route_table" {
  vpc_id = aws_vpc.vpc.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.internet_gateway.id
  }
  tags = {
    Name = "${var.tracker_id}-public-route-table"
  }
}

resource "aws_route_table_association" "public_subnet_association" {
  for_each       = aws_subnet.public_subnets
  subnet_id      = each.value.id
  route_table_id = aws_route_table.public_route_table.id
}
