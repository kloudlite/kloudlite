# kli - Kloudlite Installer CLI

A command-line tool for managing Kloudlite installations.

## Overview

`kli` is a CLI tool built with [Cobra](https://github.com/spf13/cobra) that provides an intuitive interface to create, configure, and manage Kloudlite installations from the command line.

## Installation

### Download Pre-built Binaries

Download the latest release for your platform from [GitHub Releases](https://github.com/kloudlite/kloudlite/releases?q=kli-v&expanded=true).

#### Linux (AMD64)
```bash
curl -LO https://github.com/kloudlite/kloudlite/releases/latest/download/kli-linux-amd64
chmod +x kli-linux-amd64
sudo mv kli-linux-amd64 /usr/local/bin/kli
```

#### Linux (ARM64)
```bash
curl -LO https://github.com/kloudlite/kloudlite/releases/latest/download/kli-linux-arm64
chmod +x kli-linux-arm64
sudo mv kli-linux-arm64 /usr/local/bin/kli
```

#### macOS (Intel)
```bash
curl -LO https://github.com/kloudlite/kloudlite/releases/latest/download/kli-darwin-amd64
chmod +x kli-darwin-amd64
sudo mv kli-darwin-amd64 /usr/local/bin/kli
```

#### macOS (Apple Silicon)
```bash
curl -LO https://github.com/kloudlite/kloudlite/releases/latest/download/kli-darwin-arm64
chmod +x kli-darwin-arm64
sudo mv kli-darwin-arm64 /usr/local/bin/kli
```

#### Windows
Download the appropriate `.exe` file for your architecture from the releases page and add it to your PATH.

### Build from Source

#### Quick Build
```bash
cd api/cmd/kli
task build
```

#### Build for All Platforms
```bash
cd api/cmd/kli
VERSION=0.1.0 task build-all
```

#### Install Locally
```bash
cd api/cmd/kli
task install
```

#### Run Tests
```bash
cd api/cmd/kli
task test
```

See `task --list` for all available commands.

## Usage

```bash
# Display help
kli --help
kli -h

# Show version
kli version
kli v
```

## Commands

### Version

Display the current version of the CLI:

```bash
kli version
kli v
```

### Provider Commands

Kloudlite supports installation on three major cloud providers:

#### AWS

Manage Kloudlite installations on Amazon Web Services:

```bash
# Check AWS prerequisites
kli aws doctor

# Install Kloudlite on AWS
kli aws install --installation-key prod

# Install in a specific region
kli aws install --installation-key staging --region us-west-2

# Install without termination protection (not recommended)
kli aws install --installation-key dev --enable-termination-protection=false

# Uninstall Kloudlite from AWS
kli aws uninstall --installation-key prod

# Uninstall from a specific region
kli aws uninstall --installation-key staging --region us-west-2
```

The `aws doctor` command checks:
- AWS CLI is installed
- AWS credentials are configured
- Current session has required IAM permissions

The `aws install` command:
- Requires `--installation-key` parameter to identify the installation
- Creates IAM role 'kl-{key}-role' with EC2 management permissions and SSM access
- Creates security group 'kl-{key}-sg' with required ports (443 external, 6443/8472/10250/5001 internal)
- Launches t3.medium EC2 instance 'kl-{key}-instance' with Ubuntu 24.04 LTS AMD64
- Enables AWS Systems Manager (SSM) for secure instance access without SSH keys
- Automatically installs and starts K3s server on instance startup (via cloud-init)
- Enables EC2 termination protection by default (can be disabled with `--enable-termination-protection=false`)
- Configures 100GB root volume
- Assigns public IP address
- Uses default VPC and subnet
- Tags all resources with `InstallationKey={key}` for easy identification and cleanup
- Handles interruption (Ctrl+C) gracefully by cleaning up all created resources

K3s installation details:
- Installs K3s with Traefik disabled
- Sets kubeconfig permissions to 644 for easy access
- Logs installation progress to /var/log/kloudlite-init.log
- K3s will automatically start on system boot

The `aws uninstall` command:
- Requires `--installation-key` parameter to identify which installation to remove
- Automatically disables termination protection before terminating instances
- Terminates EC2 instance(s) with the matching installation key
- Deletes security group 'kl-{key}-sg' (with automatic retries for dependency violations)
- Deletes IAM instance profile 'kl-{key}-role'
- Deletes IAM role 'kl-{key}-role' and all attached policies
- All resources are identified by the `InstallationKey` tag
- Cannot be interrupted (Ctrl+C shows warning but continues) to prevent orphaned resources

#### GCP

Manage Kloudlite installations on Google Cloud Platform:

```bash
# Check GCP prerequisites
kli gcp doctor

# Future: Install Kloudlite on GCP
kli gcp install
```

The `gcp doctor` command checks:
- gcloud CLI is installed
- gcloud is authenticated
- Default project is set
- Current session has required IAM permissions

#### Azure

Manage Kloudlite installations on Microsoft Azure:

```bash
# Check Azure prerequisites (both commands work)
kli azure doctor
kli az doctor

# Future: Install Kloudlite on Azure
kli azure install
kli az install
```

The `azure doctor` command checks:
- Azure CLI is installed
- Azure CLI is authenticated
- Default subscription is set
- Current session has required RBAC permissions

## Development

The CLI is structured following Cobra best practices:

```
cmd/kli/
├── main.go              # Entry point
├── cmd/
│   ├── root.go          # Root command with Cobra setup
│   ├── version.go       # Version command
│   ├── aws.go           # AWS provider root command
│   ├── aws_doctor.go    # AWS prerequisites check
│   ├── aws_install.go   # AWS installation command
│   ├── aws_uninstall.go # AWS uninstallation command
│   ├── gcp.go           # GCP provider root command
│   ├── gcp_doctor.go    # GCP prerequisites check
│   ├── azure.go         # Azure provider root command
│   └── azure_doctor.go  # Azure prerequisites check
└── README.md            # This file
```

## Future Commands

Additional commands will be added for each provider:
- `kli gcp install` - Install Kloudlite on GCP
- `kli azure install` - Install Kloudlite on Azure
- Configuration and setup workflows
- Status and monitoring operations

## Version

Current version: 0.1.0

## Releasing

To create a new release:

1. **Tag the release**:
   ```bash
   git tag kli-v0.1.0
   git push origin kli-v0.1.0
   ```

2. **Automated build**: The GitHub Actions workflow will automatically:
   - Build binaries for all platforms (Linux, macOS, Windows) and architectures (AMD64, ARM64)
   - Generate SHA256 checksums for verification
   - Create a GitHub release with all binaries and checksums
   - Generate installation instructions

3. **Manual trigger** (if needed):
   - Go to Actions → Release kli → Run workflow
   - Enter the tag name (e.g., `kli-v0.1.0`)

The release will be available at: `https://github.com/kloudlite/kloudlite/releases/tag/kli-v0.1.0`
