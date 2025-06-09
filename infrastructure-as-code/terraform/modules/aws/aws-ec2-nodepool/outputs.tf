output "public_ips" {
  value = {for name, result in module.ec2-nodes : name => result.public_ip}
}

#output "private_ips" {
#  value = {for name, result in module.ec2-nodes : name => result.private_ip}
#}
