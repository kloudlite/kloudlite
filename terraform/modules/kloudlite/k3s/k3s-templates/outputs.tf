output "k3s-agent-template-path" {
  value = "${path.module}/k3s-agent-setup.sh.tpl"
}

output "k3s-vm-setup-template-path" {
  value = "${path.module}/vm-setup.sh.tpl"
}

output "kloudlite_config_directory" {
  value = local.kloudlite_config_directory
}
