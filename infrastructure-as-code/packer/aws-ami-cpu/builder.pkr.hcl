packer {
  required_plugins {
    amazon = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

locals {
  ssh_username             = "ubuntu"
  k3s_version              = "v1.28.5"
  kloudlite_runner_release = "v1.0.0"
}

source "amazon-ebs" "builder" {
  ami_name        = var.dest_ami_name
  instance_type   = "t2.micro"
  source_ami      = var.source_ami
  region          = var.region
  ssh_username    = local.ssh_username
  ami_groups      = ["all"]
  ami_regions     = var.copy_to_regions
  ami_description = "kloudlite k3s (${local.k3s_version}) distribution"
  tags = {
    kloudlite-k3s-version = "v1.28.5"
  }
}

build {
  sources = [
    "sources.amazon-ebs.builder"
  ]

  provisioner "shell" {
    inline = [
      <<-EOI
      sudo cat > ~/init.sh <<EOS
      ${templatefile("../scripts/vm-init/init.sh.tpl", {
        k3s_release       = local.k3s_version
        kloudlite_release = local.kloudlite_runner_release
      })}
      EOS

      sudo bash ~/init.sh
      EOI
]
}
}

