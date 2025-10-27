import { NextRequest, NextResponse } from 'next/server';

const INSTALL_SCRIPT = `#!/bin/bash
set -e

# Colors
RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
BLUE='\\033[0;34m'
NC='\\033[0m' # No Color

# Base URL for downloads
DOWNLOAD_BASE_URL="https://console.kloudlite.io/api/download/kli"

# Detect OS and Architecture
detect_platform() {
    local os=""
    local arch=""

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        MINGW*|MSYS*|CYGWIN*) os="windows";;
        *)          echo -e "\${RED}Error: Unsupported operating system\${NC}"; exit 1;;
    esac

    # Detect Architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64";;
        arm64|aarch64)  arch="arm64";;
        *)              echo -e "\${RED}Error: Unsupported architecture: $(uname -m)\${NC}"; exit 1;;
    esac

    echo "\${os}-\${arch}"
}

# Download kli binary
download_kli() {
    local platform=$1
    local version=\${2:-latest}
    local download_url="\${DOWNLOAD_BASE_URL}/\${platform}"

    if [ "\$version" != "latest" ]; then
        download_url="\${download_url}?version=\${version}"
    fi

    echo -e "\${BLUE}Downloading kli for \${platform}...\${NC}"

    if command -v curl &> /dev/null; then
        curl -fsSL "\${download_url}" -o kli
    elif command -v wget &> /dev/null; then
        wget -q "\${download_url}" -O kli
    else
        echo -e "\${RED}Error: Neither curl nor wget found. Please install one of them.\${NC}"
        exit 1
    fi

    chmod +x kli
    echo -e "\${GREEN}✓ Downloaded successfully\${NC}"
}

# Install kli to system
install_kli() {
    local install_dir="/usr/local/bin"

    if [ -w "\$install_dir" ]; then
        mv kli "\${install_dir}/kli"
        echo -e "\${GREEN}✓ Installed to \${install_dir}/kli\${NC}"
    else
        echo -e "\${YELLOW}Installing to \${install_dir} requires sudo privileges\${NC}"
        sudo mv kli "\${install_dir}/kli"
        echo -e "\${GREEN}✓ Installed to \${install_dir}/kli\${NC}"
    fi
}

# Verify installation
verify_installation() {
    if command -v kli &> /dev/null; then
        echo -e "\${GREEN}✓ kli installed successfully!\${NC}"
        echo -e "\${BLUE}Version: \${NC}$(kli version)"
        return 0
    else
        echo -e "\${RED}✗ Installation failed\${NC}"
        return 1
    fi
}

# Print usage
print_usage() {
    cat << EOF
\${BLUE}Kloudlite Installer (kli) - Installation Script\${NC}

Usage:
  curl -fsSL https://console.kloudlite.io/api/install | bash
  curl -fsSL https://console.kloudlite.io/api/install | bash -s -- [OPTIONS]

Options:
  --version VERSION    Install specific version (default: latest)
  --help              Show this help message

Examples:
  # Install latest version
  curl -fsSL https://console.kloudlite.io/api/install | bash

  # Install specific version
  curl -fsSL https://console.kloudlite.io/api/install | bash -s -- --version 0.1.0

After installation:
  kli --help          Show kli help
  kli aws doctor      Check AWS prerequisites
  kli aws install     Install Kloudlite on AWS

EOF
}

# Main installation flow
main() {
    local version="latest"

    # Parse arguments
    while [ $# -gt 0 ]; do
        case $1 in
            --version)
                version="$2"
                shift 2
                ;;
            --help)
                print_usage
                exit 0
                ;;
            *)
                echo -e "\${RED}Unknown option: $1\${NC}"
                print_usage
                exit 1
                ;;
        esac
    done

    echo -e "\${BLUE}================================\${NC}"
    echo -e "\${BLUE}  Kloudlite Installer (kli)\${NC}"
    echo -e "\${BLUE}================================\${NC}"
    echo ""

    # Detect platform
    platform=$(detect_platform)
    echo -e "\${BLUE}Detected platform: \${platform}\${NC}"
    echo ""

    # Download
    download_kli "\$platform" "\$version"
    echo ""

    # Install
    install_kli
    echo ""

    # Verify
    if verify_installation; then
        echo ""
        echo -e "\${GREEN}🎉 Installation complete!\${NC}"
        echo ""
        echo -e "\${BLUE}Quick start:\${NC}"
        echo "  kli aws doctor        # Check AWS prerequisites"
        echo "  kli aws install       # Install Kloudlite on AWS"
        echo ""
        echo -e "\${BLUE}Documentation:\${NC}"
        echo "  https://github.com/kloudlite/kloudlite/tree/development/api/cmd/kli"
    else
        exit 1
    fi
}

main "$@"
`;

export async function GET(request: NextRequest) {
  return new NextResponse(INSTALL_SCRIPT, {
    headers: {
      'Content-Type': 'text/x-shellscript',
      'Content-Disposition': 'inline; filename="install.sh"',
      'Cache-Control': 'public, max-age=300', // Cache for 5 minutes
    },
  });
}

export const runtime = 'edge';
