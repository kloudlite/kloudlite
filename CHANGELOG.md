# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased (charts/kloudlite-agent)]

### Added

- helm chart now imports secret `kube-system/k3s-params` into the release, as it is used to provide k3s credentials to nodepool controller

## [Unreleased (charts/kloudlite-platform)]

### Added
- nodepool CRD updates, and `stateful` nodepool CR updates (due to change in CRD)
- cluster issuer, and cloudflare wildcard certificate, now honours `.Values.global.routerDomain`, when provided
- [charts/kloudlite-autoscalers] has been included as a part of this release, as it is now critical for nodes provisioning via nodepool
- when `.Values.operators.platformOperator.configuration.nodepools.extractFromCluster` is `true`, helm chart now imports secret `kube-system/k3s-params` into the release, as it is used to provide k3s credentials to nodepool controller, and which skips need to provide those cluster maintenance crdentials via helm.
- [helm/victoria-metrics] volume and storage classes for `vmselect`, and `vmcluster` have been fixed, they will now use kloudlite `ext4` (`.Values.persistence.storageClasses.ext4`) storage class, and volume size is configurable via helm values.
- [helm/nats] jetstream volume size is now configurable via helm values.
- [helm/mongodb] storage size is now configurable via helm values.
- Taskfile.yml command `debug` now uses `--dry-run=server` instead of `--dry-run`, as it allows helm to perform [`lookup`](https://helm.sh/docs/chart_template_guide/functions_and_pipelines/#using-the-lookup-function) when debugging

## [v1.0.5] - 2023-03-05
