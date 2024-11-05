locals {
  tags = merge(var.tags, { Terraform = true })
}

resource "aws_vpc" "vpc" {
  cidr_block = var.vpc_cidr
  tags       = merge(local.tags, { Name = var.vpc_name })
}

resource "aws_subnet" "public_subnets" {
  for_each                = {for idx, subnet in var.public_subnets : idx => subnet}
  vpc_id                  = aws_vpc.vpc.id
  cidr_block              = each.value.cidr
  availability_zone       = each.value.availability_zone
  map_public_ip_on_launch = true
  tags                    = merge(local.tags, {
    Name = "${var.vpc_name}-${each.value.availability_zone}-public"
  })
}

resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = aws_vpc.vpc.id
  tags   = merge(local.tags, {
    Name = "${var.vpc_name}-ingress-gateway"
  })
}

resource "aws_route_table" "public_route_table" {
  vpc_id = aws_vpc.vpc.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.internet_gateway.id
  }
  tags = merge(local.tags, {
    Name = "${var.vpc_name}-public-route-table"
  })
}

resource "aws_route_table_association" "public_subnet_association" {
  for_each       = aws_subnet.public_subnets
  subnet_id      = each.value.id
  route_table_id = aws_route_table.public_route_table.id
}
