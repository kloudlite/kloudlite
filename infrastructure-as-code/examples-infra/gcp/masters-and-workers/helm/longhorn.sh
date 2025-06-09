#! /usr/bin/env bash

cmd=$1
shift;

case $cmd in 
  install)
    helm repo add longhorn https://charts.longhorn.io
    helm repo update longhorn

    kubectl create namespace longhorn-system
    helm upgrade --install longhorn longhorn/longhorn --namespace longhorn-system -f ./values.yml
    # kubectl apply -f https://raw.githubusercontent.com/longhorn/longhorn/v1.5.1/deploy/longhorn.yaml
    ;;
  uninstall)
    helm uninstall longhorn --namespace longhorn-system
    ;;
  *)
    echo "invalid cmd ($cmd)" && exit 1
    ;;
esac
