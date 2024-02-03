terraform {
  backend "kubernetes" {
    namespace     = "default"
    secret_suffix = "state"
    config_path   = "~/.kube/configs/kloudlite-dev.yml"
  }
}
