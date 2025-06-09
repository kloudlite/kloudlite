output "vpc_id" {
  value = module.aws-vpc.vpc_id
}

output "vpc_public_subnets" {
  value = module.aws-vpc.vpc_public_subnets
}