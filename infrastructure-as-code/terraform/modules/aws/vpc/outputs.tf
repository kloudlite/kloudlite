output "vpc_id" {
  value = aws_vpc.vpc.id
}

output "vpc_public_subnets" {
  value = [
    for idx, subnet in aws_subnet.public_subnets : {
      id                = subnet.id
      availability_zone = subnet.availability_zone
      cidr              = subnet.cidr_block
    }
  ]
}
