#! /usr/bin/env bash

set -o pipefail
set -o errexit

KUBECTL="$${KUBECTL:-sudo k3s kubectl}"
KUBECONFIG="$${KUBECONFIG:-'/etc/rancher/k3s/k3s.yaml'}"
export KUBECONFIG

manifests_dir="nvidia-gpu-setup"

mkdir -p $manifests_dir

echo "[#] creating runtime class"
echo "[#]     source: https://docs.k3s.io/advanced#nvidia-container-runtime-support"
cat > $manifests_dir/runtime-class.yml <<EOF
apiVersion: node.k8s.io/v1
kind: RuntimeClass
metadata:
  name: nvidia
handler: nvidia
EOF
$KUBECTL apply -f $manifests_dir/runtime-class.yml

echo "[#] installing nvidia device plugin with helm"
echo "[#]     source: https://github.com/NVIDIA/k8s-device-plugin#deployment-via-helm"
cat > $manifests_dir/nvidia-device-plugin.yml <<EOF
apiVersion: crds.kloudlite.io/v1
kind: HelmChart
metadata:
  name: nvdp
  namespace: kube-system
spec:
  chartRepo:
    url: https://nvidia.github.io/k8s-device-plugin
    name: nvdp
  chartVersion: 0.14.1
  chartName: nvdp/nvidia-device-plugin
  jobVars:
    tolerations: ${TF_GPU_NODE_TOLERATIONS}
  valuesYaml: |+
    runtimeClassName: nvidia
    nodeSelector: ${TF_GPU_NODE_SELECTOR}
    tolerations: ${TF_GPU_NODE_TOLERATIONS}
EOF

$KUBECTL apply -f $manifests_dir/nvidia-device-plugin.yml

echo "[#] installing test-gpu pod to test nvidia runtime"
echo "[#]     source: https://docs.k3s.io/advanced#nvidia-container-runtime-support"
cat > $manifests_dir/test-gpu-pod.yml <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: nbody-gpu-benchmark
  namespace: default
spec:
  restartPolicy: OnFailure
  runtimeClassName: nvidia
  tolerations: ${TF_GPU_NODE_TOLERATIONS}
  containers:
  - name: cuda-container
    image: nvcr.io/nvidia/k8s/cuda-sample:nbody
    args: ["nbody", "-gpu", "-benchmark"]
    resources:
      limits:
        nvidia.com/gpu: 1
    env:
    - name: NVIDIA_VISIBLE_DEVICES
      value: all
    - name: NVIDIA_DRIVER_CAPABILITIES
      value: all
EOF

$KUBECTL apply -f $manifests_dir/test-gpu-pod.yml