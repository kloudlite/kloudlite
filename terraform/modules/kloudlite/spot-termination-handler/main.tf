resource "ssh_resource" "apply_spot_termination_handler" {
  host        = var.ssh_params.public_ip
  user        = var.ssh_params.username
  private_key = var.ssh_params.private_key

  timeout     = "1m"
  retry_delay = "5s"

  when = "create"

  pre_commands = [
    "mkdir -p manifests"
  ]

  file {
    content     = <<EOF
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: ${var.release_name}
  namespace: ${var.release_namespace}
spec:
  chartRepo:
    name: kloudlite
    url: https://kloudlite.github.io/helm-charts

  chartName: kloudlite/aws-spot-termination-handler
  chartVersion: ${var.kloudlite_release}

  jobVars:
    tolerations:
      - operator: "Exists"

  values:
    kloudliteRelease: ${var.kloudlite_release}
    nodeSelector: ${jsonencode(var.spot_nodes_selector)}
EOF
    destination = "manifests/spot-termination-handler.helm.yml"
    permissions = "0755"
  }

  commands = [
    <<EOC
export KUBECTL="sudo k3s kubectl"
$KUBECTL apply -f manifests/spot-termination-handler.helm.yml
EOC
  ]
}
