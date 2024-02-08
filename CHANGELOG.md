# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v1.0.5](https://github.com/kloudlite/helm-charts/compare/master...release-1.0.5) - 2024-02-01

### Added [charts/kloudlite-agent]
- helm chart now imports secret `kube-system/k3s-params` into the release, as it is used to provide k3s credentials to nodepool controller, and allow users to skip providing those values via helm.

### Added [charts/kloudlite-platform]

- image tag, now uses `.Values.global.kloudlite_release` if present otherwise `.chart.AppVersion`.
- imagePullPolicy is autoconfigured to `Always`, when image tag is like `-nightly$`, otherwise it is `IfNotPresent`.

- nodepool CRD updates, and `stateful` nodepool CR updates (due to change in CRD)
- cluster issuer, and cloudflare wildcard certificate, now honours `.Values.global.routerDomain`, when provided
- [charts/kloudlite-autoscalers] has been included as a part of this release, as it is now critical for nodes provisioning via nodepool
- when `.Values.operators.platformOperator.configuration.nodepools.extractFromCluster` is `true`, helm chart now imports secret `kube-system/k3s-params` into the release, as it is used to provide k3s credentials to nodepool controller, and which skips need to provide those cluster maintenance crdentials via helm.
- [helm/victoria-metrics] volume and storage classes for `vmselect`, and `vmcluster` have been fixed, they will now use kloudlite `ext4` (`.Values.persistence.storageClasses.ext4`) storage class, and volume size is configurable via helm values.
- [helm/nats] jetstream volume size is now configurable via helm values.
- [helm/mongodb] storage size is now configurable via helm values.

[charts/kloudlite-agent]: ./charts/kloudlite-agent
[charts/kloudlite-platform]: ./charts/kloudlite-platform
[charts/kloudlite-autoscalers]: ./charts/kloudlite-autoscalers
[crds]: ./crds
