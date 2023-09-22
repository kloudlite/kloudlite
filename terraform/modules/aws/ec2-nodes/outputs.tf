output "ec2_instances_public_ip" {
  value = {
    for name, instance in aws_instance.ec2_instances : name =>
    lookup(aws_eip.elastic_ips, name, null) != null ? aws_eip.elastic_ips[name].public_ip : instance.public_ip
  }
}

output "ec2_instances_private_ip" {
  value = {for name, instance in aws_instance.ec2_instances : name => instance.private_ip}
}

output "ec2_instances_az" {
  value = {for name, instance in aws_instance.ec2_instances : name =>  instance.availability_zone}
}

output "ssh_private_key" {
  sensitive = true
  value     = tls_private_key.ssh_key.private_key_pem
}
