variable "cluster-id" {
  default = "kl"
}

variable "do-token" {
  default = ""
}

variable "accountId" {
  default = ""
}

variable "nodeId" {
  default = ""
}

variable "size" {
  default = "s-4vcpu-8gb"
}

variable "region" {
  default = "blr1"
}

variable "keys-path" {
  # default = ""
}

# variable "pubkey" {
#   # default = ""
# }

variable "do-image-id" {
  default = "ubuntu-22-10-x64"
  # default = "105910703"
}

variable "ssh_keys" {
  default = ["25:d8:56:2b:70:15:43:c5:dd:e2:ff:d7:47:1b:68:22"]
}
