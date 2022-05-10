variable "master-node-size" {
  default = "s-4vcpu-8gb"
}

variable "master-node-data" {
  type = map(object({
    name = string
    ip = string
  }))
  default = {
    "master" : {
      name = "master",
      ip   = "10.13.13.2"
    },
    "master-1" : {
      name = "master-1",
      ip   = "10.13.13.3"
    },
    "master-2" : {
      name = "master-2",
      ip   = "10.13.13.4"
    },
  }
}

variable "master-nodes" {
  type= set(string)
  default = ["master", "master-1", "master-2"]
}

variable "agent-node-size" {
  default = "s-4vcpu-8gb"
}



variable "ssh_keys" { default = ["25:d8:56:2b:70:15:43:c5:dd:e2:ff:d7:47:1b:68:22"] }
