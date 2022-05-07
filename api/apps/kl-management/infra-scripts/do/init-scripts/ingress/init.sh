#! /usr/bin/env bash

NAMESPACE=helm-ingress
REPO_URL="https://kubernetes.github.io/ingress-nginx"
CHART_NAME=ingress-nginx
RELEASE=ingress-nginx
REPO=$CHART_NAME/ingress-nginx

[ -f ./values.yaml ] && F_VALUES='-f ./values.yaml'
[ -f ./values.yml ] && F_VALUES='-f ./values.yml'

ensureHelmRepo() {
  repoExists=$(helm repo list | grep -c "$REPO_URL")
  if [ "$repoExists" -eq 0 ]; then
    helm repo add "$CHART_NAME" "$REPO_URL"
    helm repo update
  fi
}

command=$1
shift 1

case "$command" in
  install)
    ensureHelmRepo
    helm install "$RELEASE" $REPO --create-namespace --namespace "$NAMESPACE" $F_VALUES $@
    ;;
  upgrade)
    ensureHelmRepo
    helm upgrade --install "$RELEASE" $REPO --create-namespace --namespace "$NAMESPACE" $F_VALUES $@
    ;;

  uninstall)
    ensureHelmRepo
    echo 'uninstall triggered'
    helm uninstall "$RELEASE" --namespace "$NAMESPACE"
    ;;

  *)
    echo "Usage: $0 <command>"
    echo "Commands:"
    echo "  install"
    echo "  upgrade"
    echo "  uninstall"
    exit 17
    ;;
esac

