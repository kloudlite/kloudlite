resource "aws_vpc" "vpc" {
  cidr_block = var.vpc_cidr
  tags       = merge(var.tags, {
    Name = "${var.tracker_id}-${var.vpc_name}"
  })
}

locals {
  public_subnet = {for idx, subnet in var.public_subnets : subnet.availability_zone => subnet.cidr}
}

resource "aws_subnet" "public_subnets" {
  for_each                = local.public_subnet
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = each.value
  availability_zone       = each.key
  map_public_ip_on_launch = true
  tags                    = merge(var.tags, {
    Name = "${var.tracker_id}-${each.value}-public"
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
  for_each       = local.public_subnet
  subnet_id      = aws_subnet.public_subnets[each.key].id
  route_table_id = aws_route_table.public_route_table.id
}
