output "vpc_id" {
  value = aws_vpc.vpc.id
}

output "vpc_public_subnets" {
  value = {for idx, subnet in aws_subnet.public_subnets : subnet.availability_zone => subnet.id}
}

