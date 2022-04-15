#! /usr/bin/env bash
k3s kubectl exec wireguard -n wireguard -- bash -c "/host-scripts/wgman -command=init -ip=$1"
