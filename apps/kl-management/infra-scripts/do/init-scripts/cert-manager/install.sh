#!/usr/bin/env sh

helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install \
  cert-manager jetstack/cert-manager \
  --namespace helm-cert-manager \
  --create-namespace \
  --version v1.8.0 \
  --set installCRDs=true
