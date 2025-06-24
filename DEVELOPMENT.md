# Local Development Setup

## ensure following are setup
- [nix (with flakes)] [https://determinate.systems/posts/determinate-nix-installer/]
- [nix-direnv](https://github.com/nix-community/nix-direnv)
- [docker](https://docs.docker.com/get-started/get-docker/)/[podman](https://podman.io/docs/installation)
- [docker-compose](https://docs.docker.com/compose/)/[podman-compose](https://github.com/containers/podman-compose)
- [wireguard](https://www.wireguard.com/install/)

> install mkcert TLS trust store on your machine (one time step)
> with `mkcert -install`

## start docker compose

```bash
docker-compose up -d
```

## copy kubeconfig from container to your machine

>[!NOTE]
> copy to other location if you already have ~/.kube/config

```bash
docker-compose cp k8s:/etc/rancher/k3s/k3s.yaml ~/.kube/config
```

## ensure you are able to access cluster and it's resources

```bash
kubectl get pods -A
```

## setup helm chart plugin on your local cluster

```bash
curl -L0 https://raw.githubusercontent.com/kloudlite/plugin-helm-chart/refs/heads/master/config/crd/bases/plugin-helm-chart.kloudlite.github.com_helmcharts.yaml | kubectl apply -f -
curl -L0 https://raw.githubusercontent.com/kloudlite/plugin-helm-chart/refs/heads/master/config/crd/bases/plugin-helm-chart.kloudlite.github.com_helmpipelines.yaml | kubectl apply -f -
curl -L0 https://raw.githubusercontent.com/kloudlite/plugin-helm-chart/refs/heads/master/INSTALL/k8s/setup.yaml | kubectl apply -f -

```

## create kloudlite namespace

```bash
kubectl create ns kloudlite
```

## create TLS CA secret on your cluster

```bash
kubectl create secret tls mkcert-issuer-key-pair \
--key "$(mkcert -CAROOT)"/rootCA-key.pem \
--cert "$(mkcert -CAROOT)"/rootCA.pem  \
-n kloudlite
```

## Install Helm Pipeline for local setup

```bash
go-template --values ./helm-charts/pipeline/local/values.yml ./helm-charts/pipeline/local/pipeline.yml | kubectl apply -f -
```

let it run for a minute, everything will be setup 

## grab your wireguard config

```bash
kubectl get secrets/wg-wireguard-peer1 -n wg-wireguard -o jsonpath='{.data.wg\.conf}' | base64 -d
```
