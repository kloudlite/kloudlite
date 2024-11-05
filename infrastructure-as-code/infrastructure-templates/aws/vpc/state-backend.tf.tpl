terraform {
  required_version = ">= 1.2.0"

  backend "kubernetes" {
    # read more at https://developer.hashicorp.com/terraform/language/settings/backends/kubernetes#configuration-variables
    secret_suffix = "state"

    # when running on a kubernetes cluster, specify env-vars:
    #   - KUBE_IN_CLUSTER_CONFIG="true"
    #   - KUBE_NAMESPACE="some namespace"

    # when running on local machine, uncomment the following, and pass appropriate values
    # namespace   = "default"
    # config_path = "<path-to-config>"
  }
}
