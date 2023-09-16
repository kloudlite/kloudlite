resource "helm_release" "cert_manager" {
  count = var.install_cert_manager ? 1 : 0
  name = "cert-manager"

  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"

  version          = "v1.13.0"
  namespace        = "kube-system"
  create_namespace = true

  values = []
}
