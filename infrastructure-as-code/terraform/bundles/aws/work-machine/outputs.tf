output "ssh-public-key" {
  value       = module.ssh-rsa-key.public_key
  description = "SSH public key for the ec2 node"
}
