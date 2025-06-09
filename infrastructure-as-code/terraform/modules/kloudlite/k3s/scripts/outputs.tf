output "k3s-agent-setup" {
  value = "${path.module}/k3s-agent-setup.sh.tftpl"
}

output "k3s-master-setup" {
  value = "${path.module}/k3s-master-setup.sh.tftpl"
}

output "kloudlite_config_directory" {
  value = "/etc/kloudlite"
}
