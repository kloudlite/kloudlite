#! /usr/bin/env bash

KLOUDLITE_CONFIG_DIRECTORY=/etc/kloudlite

## terraform params
K3S_SERVER_TOKEN="${k3s_server_token}"
K3S_AGENT_TOKEN="${k3s_agent_token}"
K3S_VERSION="${k3s_version}"
NODE_NAME="${node_name}"

KLOUDLITE_RELEASE="${kloudlite_release}"
BASE_DOMAIN="${base_domain}"
# --tf params:END

TOKEN=$(curl -sX PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")
INTERNAL_NODE_IP=$(curl -sH "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/local-ipv4)

debug() {
  echo "[#] $*" | tee -a "$KLOUDLITE_CONFIG_DIRECTORY/execution.log"
}

debug "ensure kloudlite config directory ($KLOUDLITE_CONFIG_DIRECTORY) exists"
mkdir -p "$KLOUDLITE_CONFIG_DIRECTORY"

debug "################# execution started at $(date) ######################"
[ $EUID -ne 0 ] && debug "this script must be run as root. current EUID is $EUID" && exit 1

create_k3s_config_file() {
  cat >"$KLOUDLITE_CONFIG_DIRECTORY/k3s.yaml" <<EOF
cluster-init: true
token: "$K3S_SERVER_TOKEN"
agent-token: "$K3S_AGENT_TOKEN"

node-name: "$NODE_NAME"

tls-san-security: true

flannel-backend: "wireguard-native"
write-kubeconfig-mode: "0644"
node-label:
  - "kloudlite.io/node.role=master"
  - "kloudlite.io/node.master.type=primary"
  - "kloudlite.io/node.ip=$INTERNAL_NODE_IP"

etcd-snapshot-compress: true
etcd-snapshot-schedule-cron: "1 2/2 * * *"

disable-helm-controller: true

disable: 
  - "traefik"

kubelet-arg:
  - "system-reserved=cpu=100m,memory=100Mi,ephemeral-storage=2Gi"
  - "kube-reserved=cpu=100m,memory=256Mi"
  - "eviction-hard=nodefs.available<5%,nodefs.inodesFree<5%,imagefs.available<5%"
EOF

  mkdir -p /etc/rancher/k3s
  ln -sf $KLOUDLITE_CONFIG_DIRECTORY/k3s.yaml /etc/rancher/k3s/config.yaml
}

install_k3s() {
  debug "installing k3s"
  export INSTALL_K3S_CHANNEL="stable"
  export INSTALL_K3S_SKIP_SELINUX_RPM="true"

  if [ -n "$K3S_VERSION" ]; then
    export INSTALL_K3S_VERSION="$K3S_VERSION"
  fi
  curl -sfL https://get.k3s.io | INSTALL_K3S_EXEC="server" sh -
  debug "k3s installed"
}

ensure_kubernetes_is_ready() {
  export KUBECTL='sudo k3s kubectl'

  echo "checking whether /etc/rancher/k3s/k3s.yaml file exists"
  while true; do
    if [ ! -f /etc/rancher/k3s/k3s.yaml ]; then
      echo 'k3s yaml not found, re-checking in 1s'
      sleep 1
      continue
    fi

    echo "/etc/rancher/k3s/k3s.yaml file found"
    break
  done

  echo "checking whether k3s server is accepting connections"
  while true; do
    lines=$($KUBECTL get nodes | wc -l)

    if [ "$lines" -lt 2 ]; then
      echo "k3s server is not accepting connections yet, retrying in 1s ..."
      sleep 1
      continue
    fi
    echo "successful, k3s server is now accepting connections"
    break
  done
}

install_helm_plugin() {
  debug "installing kloudlite helm plugin"

  curl -L0 https://raw.githubusercontent.com/kloudlite/plugin-helm-chart/refs/heads/master/config/crd/bases/plugin-helm-chart.kloudlite.github.com_helmcharts.yaml | k3s kubectl apply -f -

  curl -L0 https://raw.githubusercontent.com/kloudlite/plugin-helm-chart/refs/heads/master/deploy/k8s/setup.yaml | k3s kubectl apply -f -

  debug "installed kloudlite helm plugin"
}

install_aws_stack() {
  debug "installing aws stack helm chart"
  mkdir -p /etc/kloudlite/manifests
  pushd /etc/kloudlite/manifests
  cat >aws-stack.yml <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: kloudlite-stack
---
apiVersion: plugin-helm-chart.kloudlite.github.com/v1
kind: HelmChart
metadata:
  name: aws-stack
  namespace: kloudlite-stack
spec:
  chart:
    url: "https://kloudlite.github.io/helm-charts-extras"
    name: aws-stack
    version: v1.0.0

  jobVars:
    resources:
      cpu:
        max: 200m
        min: 200m
      memory:
        max: 400Mi
        min: 400Mi

    tolerations:
    - key: operator
      value: exists

  helmValues: {}
EOF

  k3s kubectl apply -f ./aws-stack.yml
  popd
  debug "installed aws stack helm chart"
}

install_k9s() {
  debug "installing k9s"
  USERHOME=/home/ec2-user
  mkdir -p $USERHOME/.local/bin
  mkdir -p /tmp/x
  pushd /tmp/x
  curl -L0 https://github.com/derailed/k9s/releases/download/v0.40.8/k9s_Linux_amd64.tar.gz >k9s.tar.gz
  tar xf k9s.tar.gz
  mv k9s $USERHOME/.local/bin
  popd

  rm -rf /tmp/x
  debug "installed k9s"
}

install_kloudlite_platform() {
  mkdir -p /etc/kloudlite/manifests
  pushd /etc/kloudlite/manifests
  cat >>kloudlite-platform.yml <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: kloudlite-platform
---
apiVersion: plugin-helm-chart.kloudlite.github.com/v1
kind: HelmChart
metadata:
  name: kloudlite-platform
  namespace: kloudlite-platform
spec:
  chart:
    url: https://kloudlite.github.io/helm-charts
    name: kloudlite-platform
    version: $KLOUDLITE_RELEASE
  jobVars:
    resources:
      cpu:
        max: 200m
        min: 200m
      memory:
        max: 400Mi
        min: 400Mi

    tolerations:
    - key: operator
      value: exists

  preInstall: |+
    kubectl apply -f https://github.com/kloudlite/helm-charts/releases/download/$KLOUDLITE_RELEASE/crds-all.yml --server-side
    kubectl apply -f https://raw.githubusercontent.com/kloudlite/plugin-mongodb/refs/heads/master/config/crd/bases/plugin-mongodb.kloudlite.github.com_standaloneservices.yaml
    kubectl apply -f https://raw.githubusercontent.com/kloudlite/plugin-mongodb/refs/heads/master/config/crd/bases/plugin-mongodb.kloudlite.github.com_standalonedatabases.yaml

  helmValues:
    baseDomain: "$BASE_DOMAIN"
    clusterInternalDNS: "cluster.local"
EOF

  k3s kubectl apply -f ./kloudlite-platform.yml
  popd
  debug "installed kloudlite platform"
}

create_k3s_config_file
install_k3s
ensure_kubernetes_is_ready
install_helm_plugin
install_aws_stack
install_k9s
install_kloudlite_platform

debug "################# execution finished at $(date) ######################"
