#! /usr/bin/env bash

KUBECTL="${KUBECTL:-sudo k3s kubectl}"
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
helm repo add nvdp https://nvidia.github.io/k8s-device-plugin
helm repo update nvdp

ns_nvdp="kl-nvidia-device-plugin"

helm upgrade -i nvdp nvdp/nvidia-device-plugin \
  --namespace ${ns_nvdp} \
  --create-namespace \
  --version 0.14.1 \
  --set runtimeClassName=nvidia

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