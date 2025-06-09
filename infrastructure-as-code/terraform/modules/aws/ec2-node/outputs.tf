#output "ec2_instances_public_ip" {
#  value = {
#    for name, instance in aws_instance.ec2_instances : name =>
#    lookup(aws_eip.elastic_ips, name, null) != null ? aws_eip.elastic_ips[name].public_ip : instance.public_ip
#  }
#  #  value = {
#  #    for name, instance in aws_instance.ec2_instances : name => instance.public_ip
#  #  }
#}
#
##output "ec2_instances_elastic_ips" {
##  value = {for name, instance in aws_eip.elastic_ips : name => instance.public_ip}
##}
#
#output "ec2_instances_private_ip" {
#  value = {for name, instance in aws_instance.ec2_instances : name => instance.private_ip}
#}
#
#output "ec2_instances_az" {
#  value = {for name, instance in aws_instance.ec2_instances : name =>  instance.availability_zone}
#}

#output "ssh_private_key" {
#  sensitive = true
#  value     = tls_private_key.ssh_key.private_key_pem
#}

output "public_ip" {
  value = aws_instance.ec2_instance.public_ip
}

output "private_ip" {
  value = aws_instance.ec2_instance.private_ip
}

#output "k3s_data_volume_device" {
#  value = local.k3s_data_volume_device
#}