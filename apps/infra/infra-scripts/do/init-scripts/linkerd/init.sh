#! /usr/bin/env bash

NAMESPACE=helm-linkerd
REPO_URL="https://helm.linkerd.io/stable"
REPO_NAME=linkerd
RELEASE=linkerd
CHART=$REPO_NAME/linkerd2

[ -f ./values.yaml ] && F_VALUES='-f ./values.yaml'
[ -f ./values.yml ] && F_VALUES='-f ./values.yml'

ensureHelmRepo() {
  repoExists=$(helm repo list | grep -c "$REPO_URL")
  if [ "$repoExists" -eq 0 ]; then
    helm repo add "$REPO_NAME" "$REPO_URL"
    helm repo update
  fi
}

command=$1
shift 1

exp=$(date -d '+17520 hour' +"%Y-%m-%dT%H:%M:%SZ")

case "$command" in
  install)
    ensureHelmRepo
    helm install "$RELEASE" $CHART --create-namespace --namespace "$NAMESPACE" $F_VALUES $@
    ;;
  upgrade)
    ensureHelmRepo
    helm upgrade --install "$RELEASE" $CHART --create-namespace --namespace "$NAMESPACE" $F_VALUES \
     --set-file identityTrustAnchorsPEM=ca.crt \
     --set-file identity.issuer.tls.crtPEM=issuer.crt \
     --set-file identity.issuer.tls.keyPEM=issuer.key \
     --set identity.issuer.crtExpiry="$exp" \
     $@
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

