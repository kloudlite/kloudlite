#! /usr/bin/env bash
k3s kubectl exec wireguard -n wireguard -- bash -c "echo $1 | base64 -d | /host-scripts/wgman -command=peers"
