output "ubuntu_amd64_cpu_ami_id" {
  value = data.aws_ami.ubuntu_amd64_cpu_ami.id
}

output "ubuntu_amd64_cpu_ami_ssh_username" {
  value = "ubuntu"
}

output "ubuntu_amd64_gpu_ami_id" {
  value = data.aws_ami.ubuntu_amd64_gpu_ami.id
}

output "ubuntu_amd64_gpu_ami_ssh_username" {
  value = "ubuntu"
}
