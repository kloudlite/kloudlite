output "result" {
  value = base64encode(chomp(ssh_resource.execute_command.result))
}
