# Kloudlite Platform Backend

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B37304%2Fgit%40github.com%3Akloudlite%2Fapi-go.svg?type=shield)](https://app.fossa.com/projects/custom%2B37304%2Fgit%40github.com%3Akloudlite%2Fapi-go?ref=badge_shield)
![Nightly CI](https://github.com/k3s-io/k3s/actions/workflows/nightly-install.yaml/badge.svg)
[![Build Status](https://drone-publish.k3s.io/api/badges/k3s-io/k3s/status.svg)](https://drone-publish.k3s.io/k3s-io/k3s)
[![Integration Test Coverage](https://github.com/k3s-io/k3s/actions/workflows/integration.yaml/badge.svg)](https://github.com/k3s-io/k3s/actions/workflows/integration.yaml)
[![Unit Test Coverage](https://github.com/k3s-io/k3s/actions/workflows/unitcoverage.yaml/badge.svg)](https://github.com/k3s-io/k3s/actions/workflows/unitcoverage.yaml)

Kloudlite is a cloud native platform engineering system.

## Objective:
To develop an intuitive, versatile platform focused on Infrastructure and DevOps automation, 
designed to simplify the management, setup, and maintenance of infrastructure across various 
public clouds and on-premises data-centres. This platform aims to assist developers and companies 
in streamlining their workflows, minimizing operational complexities, and improving overall 
productivity. By leveraging advanced automation technologies, the solution will enable users 
to effectively manage complex multi-cloud and hybrid infrastructures without requiring extensive 
expertise.


## Table of Contents

1. [Installation](#installation)
2. [Uninstallation](#uninstallation)
3. [Usage](#usage)
4. [Contributing](docs/code-contribution-guidelines.md)
5. [License](LICENSE)

## Installation

This section provides instructions for installing our application using Helm, a popular package manager for Kubernetes. Before proceeding, make sure you have the following prerequisites:

- Kubernetes cluster up and running
- kubectl configured to connect to your cluster
- Helm v3.x installed

### Step 1: Add the Helm Repository
First, add the Helm repository containing the chart for our application:

```
helm repo add kloudlite https://github.com/kloudlite/helm
helm repo update
```

### Step 2: Configure the Application
Create a values.yaml file to customize the application's configuration according to your needs.
Use the values.yaml file provided in the Helm chart as a reference.

```
# values.yaml
someFeature:
  enabled: true
  replicas: 2

anotherFeature:
  size: "large"
```

### Step 3: Install the Application

Install the application using the helm install command, specifying your custom values.yaml file:

```
helm install <RELEASE_NAME> <REPO_NAME>/<CHART_NAME> -f values.yaml
```

### Step 4: Verify the Installation

After the installation is complete, check that the application's resources have been created in your Kubernetes cluster:

```
kubectl get all -l app.kubernetes.io/instance=<RELEASE_NAME>
```

### Step 5: Access the Application

Depending on your application's configuration, you may need to expose its services to access it. Follow the specific instructions provided by your application's documentation.

## Uninstalling the Application

To uninstall the application, use the helm uninstall command:

```
helm uninstall <RELEASE_NAME>
```
