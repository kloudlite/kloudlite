module "constants" {
  source = "../../common/constants"
}

resource "ssh_resource" "helm_install_kloudlite_autoscalers" {
  host        = var.ssh_params.public_ip
  user        = var.ssh_params.username
  private_key = var.ssh_params.private_key

  timeout     = "1m"
  retry_delay = "5s"

  when = "create"

  pre_commands = [
    "mkdir -p manifests"
  ]

  triggers = {
    always_run = timestamp()
  }

  file {
    content     = <<EOF
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: ${var.release_name}
  namespace: ${var.release_namespace}
  labels:
    kloudlite.io/created-by: kloudlite-iac
spec:
  chartRepo:
    name: kloudlite
    url: https://kloudlite.github.io/helm-charts

  chartName: kloudlite/kloudlite-autoscalers
  chartVersion: ${var.kloudlite_release}

  jobVars:
    backOffLimit: 1
    tolerations:
      - operator: Exists

  valuesYaml: |+
    defaults:
      imageTag: ${var.kloudlite_release}
      imagePullPolicy: "Always"

    serviceAccount:
      create: true
      nameSuffix: "sa"

    clusterAutoscaler:
      enabled: true
      image:
        repository: "ghcr.io/kloudlite/cluster-autoscaler-amd64"
        tag: "kloudlite-${var.kloudlite_release}"
      tolerations:
        - operator: Exists

EOF
    destination = "manifests/helm-kloudlite-autoscalers.yml"
  }

  commands = [
    <<EOC
export KUBECTL="sudo k3s kubectl"
$KUBECTL apply -f manifests/helm-kloudlite-autoscalers.yml
EOC
  ]
}
