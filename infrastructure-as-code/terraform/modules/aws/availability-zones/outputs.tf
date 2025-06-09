output "names" {
  value = sort(data.aws_availability_zones.az.names)
}