output "ec2_instances" {
  value = {
    for name, instance in aws_instance.ec2_instances : name => {
      #      instance_id = instance.id,
      private_ip = instance.private_ip,
      public_ip  = lookup(aws_eip.elastic_ips, name, null) != null ? aws_eip.elastic_ips[name].public_ip : instance.public_ip
      #      public_ip   = has(local.nodes_with_elastic_ips, name) ? aws_eip.elastic_ips[name].public_ip : instance.public_ip,
      az         = instance.availability_zone,
    }
  }
}

output "ec2_instances_public_ip" {
  #  value = {for name, instance in aws_instance.ec2_instances : name =>  instance.public_ip}
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
  value = tls_private_key.ssh_key.private_key_pem
}
